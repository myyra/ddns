# DDNS

Simple DDNS for Cloudflare, built using these assumptions:
- All my DNS is in Cloudflare, no other provider is/will be supported unless I switch
- I want to use a reliable and well known source for the IP check (in this case Cloudflare)
- It shouldn't need any other input than the domain and subdomain the record should be updated to

## Usage

Make sure the zone has been created in Cloudflare, then run

``` shell
docker run -d --restart=always --name=ddns ghcr.io/myyra/ddns \
    -token="Your Cloudflare API token here" \
    -recordName="home.example.com" \
    -zoneName="example.com"
```

If the record for the subdomain doesn't exist yet, it will be created.

If it already exists, it will be updated only if the IP has changed.
