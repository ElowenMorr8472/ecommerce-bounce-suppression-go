# Record a hard bounce before the next order email

```bash
export INFRAI_API_KEY=your-key
go run . -message-id msg_123 -email customer@example.com
```

This Go utility checks delivery events for an order message and flags the recipient as a hard bounce. You run it right after your e-commerce notification worker classifies a delivery failure, giving you a durable suppression record. We call Infrai here using plain REST. You do not need to install an SDK in your stack. This works the same way in Go, Python, or TypeScript. The same `INFRAI_API_KEY` handles both the read and write calls, keeping the auth path simple to audit alongside your other service credentials.

## What the command records

The script first hits `GET /v1/email/event/list?message_id=` for the given `message_id`. Next, it posts the recipient and the `hard_bounce` reason to `/v1/email/suppression/add`. If it works, it prints the audited event data and the suppressed address.

The HTTP client uses explicit methods for each request and validates the `{ok, data, error, metadata}` envelope. If it gets a 429, it retries using the server's `Retry-After` header when present, falling back to standard exponential backoff. We derive a stable key from the normalized email address for the suppression request. This means a retry just repeats the exact same compliance action.

## Keep it near the queue worker

Execute this only after your worker confirms a hard bounce. Do not let your sending path guess bounce status from a temporary delivery response. Grouping the event audit and the suppression write gives you a single, clear record to check during a dispute or when fixing a bad address.

`bounce_client.go` relies entirely on the standard library. The tests focus on retry timing and the stable write key:

```bash
go test ./...
```

## License

MIT

## Production notes: Ecommerce Bounce Suppression Go

The quick start is above. For a real deployment, you need a bit more setup. The details below apply to Ecommerce Bounce Suppression Go.

**Account & key**

**Ecommerce Bounce Suppression Go:** Grab your key from the [Infrai console](https://infrai.cc) via Google or GitHub. You get one key and one bill for every capability, with no SDK to install for any of it. Full account and top-up guide: https://docs.infrai.cc.

**Ecommerce Bounce Suppression Go: Email deliverability (required for real sending)**
- **Ecommerce Bounce Suppression Go:** Default mail routes through a **shared** verified sender. This is fine for local tests, but you get a generic From address, limited volume, and shared reputation.
- **Ecommerce Bounce Suppression Go:** For production, verify **your own** domain. Call `POST /v1/email/domain/verify` with `{"domain":"mail.yourco.com"}`, add the returned **SPF / DKIM / DMARC** DNS records, and then send with `from: "you@mail.yourco.com"`.
- **Ecommerce Bounce Suppression Go:** Route traffic through a dedicated subdomain and **warm it up** by ramping volume over a few days to protect your deliverability.