package main

import (
	"fmt"
	"github.com/gocolly/colly"
	"github.com/robfig/cron/v3"
	mail "github.com/xhit/go-simple-mail/v2"
	urlUtils "net/url"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
)

type listing struct{ title, price, url string }
type difference struct{ added, removed, updatedOld, updatedNew []listing }

func getDifference(old, new map[string]listing) difference {
	diff := difference{}

	for url, item := range old {
		if _, ok := new[url]; !ok {
			diff.removed = append(diff.removed, item)
		} else if newItem := new[url]; newItem != item {
			diff.updatedOld = append(diff.updatedOld, item)
			diff.updatedNew = append(diff.updatedNew, newItem)
		}
	}

	for url, item := range new {
		if _, ok := old[url]; !ok {
			diff.added = append(diff.added, item)
		}
	}

	return diff
}

func buildMessage(diff difference) string {
	var generateSection = func(html *strings.Builder, title string, listings []listing) {
		if len(listings) > 0 {
			html.WriteString(fmt.Sprintf("<h1>%s:</h1>", title))
			html.WriteString("<table border='1'><tr><th>Title</th><th>Price</th></tr>")
			for _, l := range listings {
				html.WriteString("<tr>")
				html.WriteString(fmt.Sprintf("<td><a href='%s'>%s</a></td><td>%s</td>", l.url, l.title, l.price))
				html.WriteString("</tr>")
			}
			html.WriteString("</table>")
		}
	}

	var html strings.Builder

	html.WriteString("<html><head><title>Listings Difference</title></head><body>")
	generateSection(&html, "Added Listings", diff.added)
	generateSection(&html, "Updated Listings New", diff.updatedNew)
	generateSection(&html, "Updated Listings Old", diff.updatedOld)
	html.WriteString("</body></html>")

	return html.String()
}

func sendMail(message, smtpHost, smtpUser, smtpPass, from, to, subject string, smtpPort int) {
	server := mail.NewSMTPClient()

	server.Host = smtpHost
	server.Port = smtpPort
	server.Username = smtpUser
	server.Password = smtpPass
	server.Encryption = mail.EncryptionSTARTTLS
	server.Authentication = mail.AuthLogin

	client, _ := server.Connect()

	email := mail.NewMSG()
	email.SetFrom(from).
		AddTo(to).
		SetSubject(subject).
		SetBody(mail.TextHTML, message)

	_ = email.Send(client)
}

func scrapListings(url string, pageLimit int) map[string]listing {
	var sanitizePrice = func(input string) string {
		re := regexp.MustCompile(`(\d{1,3}(?:\s\d{3})*) zł`)
		match := re.FindStringSubmatch(input)

		if len(match) > 0 {
			return match[0]
		} else {
			return "Check by yourself :-)"
		}
	}

	parsedURL, _ := urlUtils.Parse(url)
	baseUrl := parsedURL.Scheme + "://" + parsedURL.Hostname()
	scrapedListings := make(map[string]listing)
	pageScraped := 0
	nextPage := baseUrl

	c := colly.NewCollector()

	c.OnHTML("div[data-testid=\"pagination-wrapper\"]", func(e *colly.HTMLElement) {
		nextPage = baseUrl + e.ChildAttr("a[data-testid=\"pagination-forward\"]", "href")
	})

	c.OnHTML("div[data-testid=\"l-card\"]", func(e *colly.HTMLElement) {
		scrapedListing := listing{}

		scrapedListing.title = e.ChildText("h6")
		scrapedListing.price = sanitizePrice(e.ChildText("p[data-testid=\"ad-price\"]"))
		scrapedListing.url = baseUrl + e.ChildAttr("a", "href")

		scrapedListings[scrapedListing.url] = scrapedListing
	})

	c.OnScraped(func(response *colly.Response) {
		pageScraped++
		if nextPage != baseUrl && pageScraped < pageLimit {
			_ = c.Visit(nextPage)
		}
	})

	_ = c.Visit(url)

	return scrapedListings
}

func main() {
	url := os.Getenv("URL")
	cronExp := os.Getenv("CRON_EXP")
	pageLimit, _ := strconv.Atoi(os.Getenv("PAGE_LIMIT"))
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	mailFrom := os.Getenv("MAIL_FROM")
	mailTo := os.Getenv("MAIL_TO")
	mailSubj := os.Getenv("MAIL_SUBJ")

	fmt.Println("Loaded url:", url)
	fmt.Println("Loaded cron expression:", cronExp)
	fmt.Println("Loaded page limit:", pageLimit)
	fmt.Println("Loaded smtp host:", smtpHost)
	fmt.Println("Loaded smtp port:", smtpPort)
	fmt.Println("Loaded smtp user:", smtpUser)
	if len(smtpPass) > 0 {
		fmt.Println("Loaded smtp password: ****")
	}
	fmt.Println("Loaded email from:", mailFrom)
	fmt.Println("Loaded email to:", mailTo)
	fmt.Println("Loaded email subject:", mailSubj)

	c := cron.New()

	listings := scrapListings(url, pageLimit)
	fmt.Println("Initialized default listing state! Found", len(listings), "results!")

	_, _ = c.AddFunc(cronExp, func() {
		newListings := scrapListings(url, pageLimit)

		if len(newListings) == 0 {
			fmt.Println("Something went wrong. skipping listing retrieval.")
			return
		}

		diff := getDifference(listings, newListings)

		if len(diff.removed) > 0 {
			fmt.Println("Removed", len(diff.removed), "listings.")
		}

		if len(diff.added) > 0 || len(diff.updatedNew) > 0 {
			fmt.Println("Found difference!", len(diff.added), "added,", len(diff.updatedNew), "updated. Sending email.")
			message := buildMessage(diff)
			sendMail(message, smtpHost, smtpUser, smtpPass, mailFrom, mailTo, mailSubj, smtpPort)
			fmt.Println("Email sent successfully!")
		} else {
			fmt.Println("No difference found.")
		}

		listings = newListings
	})

	c.Start()
	fmt.Println("Successfully started scouting.")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	fmt.Println("Press Ctrl+C to exit")
	<-sigCh

	fmt.Println("Exiting.")
}
