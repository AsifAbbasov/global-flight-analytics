# Document 217 — readsb-Inspired Position Source Provenance

## Status

R2 engineering implementation on a stacked feature branch. This document does not authorize merge.

## Goal

Preserve the position-method evidence already present in readsb-compatible provider payloads and expose it separately from the provider/feed identity.

The project must never infer a position method from a provider brand. An observation supplied by ADSB.lol may be ADS-B, MLAT, ADS-R, TIS-B, ADS-C, or unavailable.

## Canonical mapping

| readsb type | GFA position_source |
| --- | --- |
| adsb_icao, adsb_icao_nt, adsb_other | adsb |
| adsr_icao, adsr_other | adsr |
| tisb_icao, tisb_other, tisb_trackfile | tisb |
| adsc | adsc |
| mlat | mlat |
| mode_s, other, empty, unknown future values | unavailable |

Mode-S is not promoted to a position method because the readsb contract describes it as transponder data without a transmitted position.

## Evidence separation

```text
source_name      = provider/feed identity
position_source  = observed position method
```

Examples:

```text
source_name=adsb.lol
position_source=mlat
```

is valid and must not be rewritten to ADS-B.

## Persistence

Migration 031 expands the existing flight_states.position_source constraint with adsr, tisb, and adsc. Existing adsb, asterix, mlat, flarm, and unavailable values remain valid. No historical row is rewritten.

## API

Current Traffic exposes:

- position_source: nullable canonical method;
- source_name: provider/feed identity.

Null position_source means evidence is unavailable. It is not a negative quality judgment.

## Frontend

Aircraft Detail displays a Position provenance section with:

- Method;
- Data feed;
- a method-specific factual explanation.

No synthetic confidence score, accuracy percentage, provider ranking, or preferred-source claim is introduced.

## Cost boundary

- new provider: NO
- new external call: NO
- new service: NO
- Redis/Valkey: NO
- Python runtime: NO
- paid infrastructure: NO
- hardware requirement: NO
- additional monetary cost: 0 RUB

## Stack dependency

R2 is based on R1 Position Freshness and therefore cannot become a canonical main merge candidate until PR #174 and R1 are merged and R2 is reconciled onto the resulting main. Any reconciliation changes the exact head and requires fresh validation and fresh merge authorization.
