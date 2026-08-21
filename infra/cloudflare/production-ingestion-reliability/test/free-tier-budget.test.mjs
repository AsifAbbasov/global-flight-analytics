import assert from 'node:assert/strict'
import fs from 'node:fs'
import test from 'node:test'

const wrangler = fs.readFileSync(
  'infra/cloudflare/production-ingestion-reliability/wrangler.jsonc',
  'utf8',
)
const readme = fs.readFileSync(
  'infra/cloudflare/production-ingestion-reliability/README.md',
  'utf8',
)

test('Cloudflare production reliability profile is free-tier bounded', () => {
  assert.match(wrangler, /"17,47 \* \* \* \*"/)
  assert.match(wrangler, /"19 \*\/2 \* \* \*"/)
  assert.match(wrangler, /"PRIMARY_CRON": "17,47 \* \* \* \*"/)
  assert.match(wrangler, /"WATCHDOG_CRON": "19 \*\/2 \* \* \*"/)
  assert.match(wrangler, /"DISPATCH_ENABLED": "false"/)

  assert.doesNotMatch(wrangler, /3,13,23,33,43,53/)
  assert.doesNotMatch(wrangler, /"\*\/5 \* \* \* \*"/)

  assert.match(readme, /primary Cron Trigger requests a GitHub workflow dispatch every 30 minutes/)
  assert.match(readme, /watchdog checks public traffic freshness every two hours/)
  assert.match(readme, /must remain dispatch\/manual-owned/)
})
