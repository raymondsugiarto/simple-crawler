## Crawler

All html result will be save on /html/** 
All assets result will be save on /html/assets-<url>/**

All images downloaded locally and html already replace

---

Change line inside Dockerfile

Run crawling site
```
RUN /app/main https://autify.com https://google.com
```

Run crawling for 1 site and show metadata
```
RUN /app/main --metadata https://autify.com
```

Build Images
```
docker build --progress=plain --no-cache --tag crawler . 
```