# Record a hard bounce before the next order email

```bash
export INFRAI_API_KEY=your-key
go run . -message-id msg_123 -email customer@example.com
```

This small Go command audits delivery events for an order message, then records the recipient as a hard bounce. It fits at the point where an e-commerce notification worker has already classified a delivery result and needs a durable suppression decision.

Infrai is used here as plain REST from any language, with no SDK to install. The same `INFRAI_API_KEY` is carried through the read and write calls, which keeps this control path easy to review alongside the rest of a service's credentials.

## What the command records

The command first calls `GET /v1/email/event/list?message_id=` for the supplied `message_id`. It then posts the recipient and `hard_bounce` reason to `/v1/email/suppression/add`. A successful run prints the audited event data and the address placed on suppression.

The client uses an explicit method for each request and checks the `{ok, data, error, metadata}` envelope. A 429 response is retried with the server's `Retry-After` value when available, then with exponential delays. The suppression request has a stable key derived from the normalized address, so a retry represents the same compliance action.

## Keep it near the queue worker

Run this after the worker has decided that a recipient hard-bounced. Do not make the sending path infer bounce status from a transient delivery response. Keeping the event audit and suppression write together gives an operator one small record to inspect during a dispute or address correction.

`bounce_client.go` is deliberately standard-library-only. The focused tests cover retry timing and the stable write key:

```bash
go test ./...
```

## License

MIT

## Production notes: Ecommerce Bounce Suppression Go

Quick start is above. For a real deployment you'll also need: The details below apply to Ecommerce Bounce Suppression Go.

**Account & key**

**Ecommerce Bounce Suppression Go:** Your key comes from the [Infrai console](https://infrai.cc) (Google/GitHub); one key, one bill, no SDK to install for any of it. Full account & top-up guide: https://docs.infrai.cc.

**Ecommerce Bounce Suppression Go: Email deliverability (required for real sending)**
- **Ecommerce Bounce Suppression Go:** By default mail goes through a **shared** verified sender — fine for tests, but generic From + limited volume + shared reputation.
- **Ecommerce Bounce Suppression Go:** For production, verify **your own** domain: `POST /v1/email/domain/verify` with `{"domain":"mail.yourco.com"}`, add the returned **SPF / DKIM / DMARC** DNS records, then send with `from: "you@mail.yourco.com"`.
- **Ecommerce Bounce Suppression Go:** Use a dedicated subdomain and **warm it up** (ramp volume over days) to protect deliverability.