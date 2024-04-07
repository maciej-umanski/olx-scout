# OLX Scout
A basic application written in Go that does some cron-scheduled web scraping with email notifications.

## Disclaimer
This application is provided as-is, with no guarantees whatsoever. Probably no further improvements will be done.
It might work, it might not. It might scrape listings, it might just make your computer explode.
Use at your own risk. I take no responsibility for anything that happens as a result of using this app.

**There is no error handling, no input validation also there is minimal data consistency assurance.
It works quite reliably for up to ~150 results, after that it will probably spam your inbox with nonsense.
Way to overcome this limitation is to use "Sorting by Newest" when applying filters on the website.**

## How to Run
### Prerequisites
* Dependent of interesting runtime option: [Docker](https://www.docker.com) or [Go](https://go.dev) Installed
* SMTP Server: do you own research, I use free [mailtrap](https://mailtrap.io) account. Basic configuration is already there.
* Understanding of required variables:
   * **CRON_EXP** - Cron expression for scheduling listings retrieval. Read more [here](https://crontab.guru).
   * **URL** - Url of a specified olx search. Just copy the url from the browser after applying interesting filters.
   * **PAGE_LIMIT** - Some searches are hundreds of pages long. You can limit the search to a specific value.
   * **SMTP_*** - Emailing Server configuration, basic settings for mailtrap is there, bring your own password.
   * **MAIL_TO** - Target email address. This is where the emails will be sent to.
   * **MAIL_FROM** - Source email address. **May be somehow restricted by smtp server provider.**
   * **MAIL_SUBJ** - Just a subject for an update email.

### Executable (Unix)
1. Install the required dependencies with `go mod download`
2. Build the app using `go build scout.go`
3. Create and fill env file `cp .env.example .env`
4. (Optional) Copy files `.env`, `run`, `scout` somewhere else.
5. Run the app using `./run`

### Docker
1. Build image with `docker build . -t scout:0.0.1`
2. Run the container using following command, fill up environment variables first.
   ```shell
    docker run -d \
    -e CRON_EXP='* * * * *' \
    -e URL='https://www.olx.pl/motoryzacja/samochody/bmw/?search%5Border%5D=created_at:desc&search%5Bfilter_float_price:to%5D=5500&search%5Bfilter_enum_model%5D%5B0%5D=3-as-sorozat&search%5Bfilter_enum_petrol%5D%5B0%5D=petrol&search%5Bfilter_enum_condition%5D%5B0%5D=notdamaged&search%5Bfilter_enum_transmission%5D%5B0%5D=manual' \
    -e PAGE_LIMIT='5' \
    -e SMTP_HOST='live.smtp.mailtrap.io' \
    -e SMTP_PORT='587' \
    -e SMTP_USER='api' \
    -e SMTP_PASS='' \
    -e MAIL_TO='' \
    -e MAIL_FROM='test@demomailtrap.com' \
    -e MAIL_SUBJ='Scout update!' \
    scout:0.0.1
   ```

Happy scraping! Or not.
