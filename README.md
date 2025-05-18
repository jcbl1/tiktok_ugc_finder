A project that helps user retrieve UGC creators infomation from TikTok based on hashtags provided. UGC, User-Generated Content, is a term in modern social media like Instagram and TikTok, etc, basically referring those who create content themselves and sometimes embed commercial ads in their content. This is a popular way of doing business nowadays. This project helps UGC managers (what I was told to call the person who asked me to make this project) find UGC creators they need and saves a ton of time.

## Tech Stack

- [Golang](https://go.dev/)
- [Chromedp](https://github.com/chromedp/chromedp)
- Basic knowledge of HTML&CSS&Javascript&XPath

## Prerequisites

Follow the steps in [tiktok-hashtag-analysis](https://github.com/bellingcat/tiktok-hashtag-analysis).

## TODO

- [x] subcammand "mend": fix "0" AP and AI in the result file
- [x] bug fixing: none-video links will cause API server Internal error, so links should be checked before request.
