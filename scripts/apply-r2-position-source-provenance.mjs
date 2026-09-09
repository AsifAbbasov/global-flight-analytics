#!/usr/bin/env node
import fs from 'node:fs'
import { execFileSync } from 'node:child_process'

function read(path) {
  return fs.readFileSync(path, 'utf8')
}

function write(path, content) {
  fs.mkdirSync(path.split('/').slice(0, -1).join('/'), { recursive: true })
  fs.writeFileSync(path, content)
}

function replaceExact(path, from, to, expectedCount = 1) {
  const current = read(path)
  const count = current.split(from).length - 1
  if (count !== expectedCount) {
    throw new Error(`${path}: expected ${expectedCount} matches, found ${count}`)
  }
  write(path, current.split(from).join(to))
}

function ensureAbsent(path, needle) {
  if (read(path).includes(needle)) {
    throw new Error(`${path}: already contains ${needle}`)
  }
}

const observationMetadata = 'apps/api/internal/domain/flightstate/observation_metadata.go'
replaceExact(
  observationMetadata,
  `const (\n\tPositionSourceUnknown PositionSource = ""\n\tPositionSourceADSB    PositionSource = "adsb"\n\tPositionSourceASTERIX PositionSource = "asterix"\n\tPositionSourceMLAT    PositionSource = "mlat"\n\tPositionSourceFLARM   PositionSource = "flarm"\n)`,
  `const (\n\tPositionSourceUnknown PositionSource = ""\n\tPositionSourceADSB    PositionSource = "adsb"\n\tPositionSourceADSR    PositionSource = "adsr"\n\tPositionSourceTISB    PositionSource = "tisb"\n\tPositionSourceADSC    PositionSource = "adsc"\n\tPositionSourceASTERIX PositionSource = "asterix"\n\tPositionSourceMLAT    PositionSource = "mlat"\n\tPositionSourceFLARM   PositionSource = "flarm"\n)`,
)
replaceExact(
  observationMetadata,
  `\tcase PositionSourceUnknown,\n\t\tPositionSourceADSB,\n\t\tPositionSourceASTERIX,\n\t\tPositionSourceMLAT,\n\t\tPositionSourceFLARM:`,
  `\tcase PositionSourceUnknown,\n\t\tPositionSourceADSB,\n\t\tPositionSourceADSR,\n\t\tPositionSourceTISB,\n\t\tPositionSourceADSC,\n\t\tPositionSourceASTERIX,\n\t\tPositionSourceMLAT,\n\t\tPositionSourceFLARM:`,
)

const readsbMapper = 'apps/api/internal/integrations/readsbcompat/mapper.go'
ensureAbsent(readsbMapper, 'func PositionSource(')
replaceExact(
  readsbMapper,
  `func MapAircraft(\n`,
  `func PositionSource(\n\tvalue string,\n) flightstate.PositionSource {\n\tswitch strings.ToLower(strings.TrimSpace(value)) {\n\tcase "adsb_icao", "adsb_icao_nt", "adsb_other":\n\t\treturn flightstate.PositionSourceADSB\n\tcase "adsr_icao", "adsr_other":\n\t\treturn flightstate.PositionSourceADSR\n\tcase "tisb_icao", "tisb_other", "tisb_trackfile":\n\t\treturn flightstate.PositionSourceTISB\n\tcase "adsc":\n\t\treturn flightstate.PositionSourceADSC\n\tcase "mlat":\n\t\treturn flightstate.PositionSourceMLAT\n\tdefault:\n\t\treturn flightstate.PositionSourceUnknown\n\t}\n}\n\nfunc MapAircraft(\n`,
)
replaceExact(
  readsbMapper,
  `\t\tSquawkCode:                 strings.TrimSpace(item.Squawk),\n`,
  `\t\tSquawkCode:                 strings.TrimSpace(item.Squawk),\n\t\tPositionSource:             PositionSource(item.Type),\n`,
)

write(
  'apps/api/internal/domain/flightstate/position_source_test.go',
  `package flightstate\n\nimport "testing"\n\nfunc TestNormalizePositionSourceAcceptsCanonicalValues(t *testing.T) {\n\tvalues := []PositionSource{\n\t\tPositionSourceUnknown,\n\t\tPositionSourceADSB,\n\t\tPositionSourceADSR,\n\t\tPositionSourceTISB,\n\t\tPositionSourceADSC,\n\t\tPositionSourceASTERIX,\n\t\tPositionSourceMLAT,\n\t\tPositionSourceFLARM,\n\t}\n\n\tfor _, value := range values {\n\t\tnormalized, err := NormalizePositionSource(value)\n\t\tif err != nil {\n\t\t\tt.Fatalf("normalize %q: %v", value, err)\n\t\t}\n\t\tif normalized != value {\n\t\t\tt.Fatalf("normalize %q = %q", value, normalized)\n\t\t}\n\t}\n}\n\nfunc TestNormalizePositionSourceRejectsNonCanonicalReadsbType(t *testing.T) {\n\tif _, err := NormalizePositionSource("adsb_icao"); err == nil {\n\t\tt.Fatal("raw readsb type must be normalized at the integration boundary")\n\t}\n}\n`,
)

write(
  'apps/api/internal/integrations/readsbcompat/position_source_test.go',
  `package readsbcompat\n\nimport (\n\t"testing"\n\t"time"\n\n\t"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"\n)\n\nfunc TestPositionSourceMapsReadsbEvidence(t *testing.T) {\n\ttests := []struct {\n\t\tname string\n\t\traw  string\n\t\twant flightstate.PositionSource\n\t}{\n\t\t{name: "adsb icao", raw: "adsb_icao", want: flightstate.PositionSourceADSB},\n\t\t{name: "adsb non transponder", raw: "adsb_icao_nt", want: flightstate.PositionSourceADSB},\n\t\t{name: "adsb other", raw: "adsb_other", want: flightstate.PositionSourceADSB},\n\t\t{name: "adsr icao", raw: "adsr_icao", want: flightstate.PositionSourceADSR},\n\t\t{name: "adsr other", raw: "adsr_other", want: flightstate.PositionSourceADSR},\n\t\t{name: "tisb icao", raw: "tisb_icao", want: flightstate.PositionSourceTISB},\n\t\t{name: "tisb other", raw: "tisb_other", want: flightstate.PositionSourceTISB},\n\t\t{name: "tisb trackfile", raw: "tisb_trackfile", want: flightstate.PositionSourceTISB},\n\t\t{name: "adsc", raw: "adsc", want: flightstate.PositionSourceADSC},\n\t\t{name: "mlat", raw: "mlat", want: flightstate.PositionSourceMLAT},\n\t\t{name: "mode s has no position method", raw: "mode_s", want: flightstate.PositionSourceUnknown},\n\t\t{name: "other quality unknown", raw: "other", want: flightstate.PositionSourceUnknown},\n\t\t{name: "unrecognized", raw: "future_type", want: flightstate.PositionSourceUnknown},\n\t\t{name: "empty", raw: "", want: flightstate.PositionSourceUnknown},\n\t}\n\n\tfor _, test := range tests {\n\t\tt.Run(test.name, func(t *testing.T) {\n\t\t\tif got := PositionSource(test.raw); got != test.want {\n\t\t\t\tt.Fatalf("PositionSource(%q) = %q, want %q", test.raw, got, test.want)\n\t\t\t}\n\t\t})\n\t}\n}\n\nfunc TestMapAircraftPreservesPositionSourceEvidence(t *testing.T) {\n\tmapped := MapAircraft(\n\t\t"adsb.lol",\n\t\tAircraftItem{\n\t\t\tHex:       "4b1801",\n\t\t\tLatitude:  40.4,\n\t\t\tLongitude: 49.8,\n\t\t\tType:      "mlat",\n\t\t},\n\t\ttime.Date(2026, time.September, 9, 12, 0, 0, 0, time.UTC),\n\t)\n\n\tif mapped.PositionSource != flightstate.PositionSourceMLAT {\n\t\tt.Fatalf("position source = %q, want %q", mapped.PositionSource, flightstate.PositionSourceMLAT)\n\t}\n\tif mapped.SourceName != "adsb.lol" {\n\t\tt.Fatalf("source name = %q, want adsb.lol", mapped.SourceName)\n\t}\n}\n`,
)

write(
  'database/migrations/031_expand_flight_state_position_sources.sql',
  `BEGIN;\n\nALTER TABLE flight_states\n    DROP CONSTRAINT flight_states_position_source_check;\n\nALTER TABLE flight_states\n    ADD CONSTRAINT flight_states_position_source_check\n        CHECK (\n            position_source IN (\n                '',\n                'adsb',\n                'adsr',\n                'tisb',\n                'adsc',\n                'asterix',\n                'mlat',\n                'flarm'\n            )\n        );\n\nCOMMENT ON COLUMN flight_states.position_source IS\n    'Canonical observed position method. Empty means unavailable; provider/feed identity remains in source_name.';\n\nCOMMIT;\n`,
)

const catalog = 'apps/api/internal/database/migrationfile/production_catalog_regression_test.go'
replaceExact(
  catalog,
  `\t\t30: "030_add_flight_state_message_observation_time.sql",\n`,
  `\t\t30: "030_add_flight_state_message_observation_time.sql",\n\t\t31: "031_expand_flight_state_position_sources.sql",\n`,
)
replaceExact(catalog, 'if len(orderedVersions) != 30 {', 'if len(orderedVersions) != 31 {')
replaceExact(
  catalog,
  '"production migration count = %d, want 30 (%s)"',
  '"production migration count = %d, want 31 (%s)"',
)

write(
  'apps/api/internal/repository/postgres/position_source_migration_contract_test.go',
  `package postgres\n\nimport (\n\t"os"\n\t"path/filepath"\n\t"runtime"\n\t"strings"\n\t"testing"\n)\n\nfunc TestPositionSourceMigrationExpandsEvidenceWithoutRewritingHistory(t *testing.T) {\n\t_, currentFile, _, ok := runtime.Caller(0)\n\tif !ok {\n\t\tt.Fatal("resolve current file")\n\t}\n\tpath := filepath.Clean(filepath.Join(\n\t\tfilepath.Dir(currentFile),\n\t\t"../../../../../database/migrations/031_expand_flight_state_position_sources.sql",\n\t))\n\tcontent, err := os.ReadFile(path)\n\tif err != nil {\n\t\tt.Fatalf("read migration: %v", err)\n\t}\n\ttext := string(content)\n\tfor _, required := range []string{\n\t\t"DROP CONSTRAINT flight_states_position_source_check",\n\t\t"'adsb'",\n\t\t"'adsr'",\n\t\t"'tisb'",\n\t\t"'adsc'",\n\t\t"'asterix'",\n\t\t"'mlat'",\n\t\t"'flarm'",\n\t\t"provider/feed identity remains in source_name",\n\t} {\n\t\tif !strings.Contains(text, required) {\n\t\t\tt.Fatalf("migration missing %q", required)\n\t\t}\n\t}\n\tif strings.Contains(text, "UPDATE flight_states") {\n\t\tt.Fatal("migration must not rewrite historical position-source evidence")\n\t}\n}\n`,
)

const trafficModel = 'apps/api/internal/domain/traffic/model.go'
replaceExact(
  trafficModel,
  `type CurrentTrafficItem struct {\n\tICAO24            string\n\tCallsign          string\n\tLatitude          float64\n\tLongitude         float64\n\tAltitudeM         *float64\n\tAltitudeStatus    flightstate.AltitudeStatus\n\tAltitudeSource    AltitudeSource\n\tVelocityMPS       float64\n\tHeadingDegrees    float64\n\tOnGround          bool\n\tObservedAt        time.Time\n\tMessageObservedAt *time.Time\n\tAircraftModel     string\n\tAirline           string\n\tOriginCountry     string\n}`,
  `type CurrentTrafficItem struct {\n\tICAO24            string\n\tCallsign          string\n\tLatitude          float64\n\tLongitude         float64\n\tAltitudeM         *float64\n\tAltitudeStatus    flightstate.AltitudeStatus\n\tAltitudeSource    AltitudeSource\n\tVelocityMPS       float64\n\tHeadingDegrees    float64\n\tOnGround          bool\n\tObservedAt        time.Time\n\tMessageObservedAt *time.Time\n\tPositionSource    flightstate.PositionSource\n\tSourceName        string\n\tAircraftModel     string\n\tAirline           string\n\tOriginCountry     string\n}`,
)

const trafficRepo = 'apps/api/internal/repository/postgres/traffic_repository.go'
replaceExact(
  trafficRepo,
  `\t\t\tfs.observed_at,\n\t\t\tfs.message_observed_at,\n\t\t\tCOALESCE(am.model, ''),`,
  `\t\t\tfs.observed_at,\n\t\t\tfs.message_observed_at,\n\t\t\tfs.position_source,\n\t\t\tCOALESCE(fs.source_name, ''),\n\t\t\tCOALESCE(am.model, ''),`,
  2,
)
replaceExact(
  trafficRepo,
  `\t\tvar messageObservedAt pgtype.Timestamptz\n`,
  `\t\tvar messageObservedAt pgtype.Timestamptz\n\t\tvar positionSource string\n`,
)
replaceExact(
  trafficRepo,
  `\t\t\t&item.ObservedAt,\n\t\t\t&messageObservedAt,\n\t\t\t&item.AircraftModel,`,
  `\t\t\t&item.ObservedAt,\n\t\t\t&messageObservedAt,\n\t\t\t&positionSource,\n\t\t\t&item.SourceName,\n\t\t\t&item.AircraftModel,`,
)
replaceExact(
  trafficRepo,
  `\t\tif messageObservedAt.Valid {\n\t\t\tvalue := messageObservedAt.Time.UTC()\n\t\t\titem.MessageObservedAt = &value\n\t\t}\n\n\t\titem.AltitudeM,`,
  `\t\tif messageObservedAt.Valid {\n\t\t\tvalue := messageObservedAt.Time.UTC()\n\t\t\titem.MessageObservedAt = &value\n\t\t}\n\n\t\tnormalizedPositionSource, err := flightstate.NormalizePositionSource(\n\t\t\tflightstate.PositionSource(positionSource),\n\t\t)\n\t\tif err != nil {\n\t\t\treturn nil, err\n\t\t}\n\t\titem.PositionSource = normalizedPositionSource\n\n\t\titem.AltitudeM,`,
)

const altitudeFixture = 'apps/api/internal/repository/postgres/traffic_altitude_semantics_integration_test.go'
replaceExact(
  altitudeFixture,
  `\t\t\t\tmessage_observed_at timestamptz,\n\t\t\t\torigin_country text`,
  `\t\t\t\tmessage_observed_at timestamptz,\n\t\t\t\tposition_source text NOT NULL DEFAULT '',\n\t\t\t\tsource_name text NOT NULL DEFAULT '',\n\t\t\t\torigin_country text`,
)

write(
  'apps/api/internal/repository/postgres/traffic_position_provenance_integration_test.go',
  `package postgres\n\nimport (\n\t"context"\n\t"testing"\n\t"time"\n\n\t"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"\n)\n\nfunc TestTrafficRepositoryPreservesPositionAndFeedProvenance(t *testing.T) {\n\tfixture := newTrafficAltitudeFixture(t)\n\tctx := context.Background()\n\trunID := "22222222-2222-2222-2222-222222222222"\n\tobservedAt := time.Date(2026, time.September, 9, 12, 0, 0, 0, time.UTC)\n\n\tmustExecTrafficAltitudeSQL(\n\t\tt,\n\t\tfixture.pool,\n\t\t\`\n\t\t\tINSERT INTO ingestion_runs (id, finished_at, status, created_at)\n\t\t\tVALUES ($1, $2, 'success', $2)\n\t\t\`,\n\t\trunID,\n\t\tobservedAt,\n\t)\n\n\tmustExecTrafficAltitudeSQL(\n\t\tt,\n\t\tfixture.pool,\n\t\t\`\n\t\t\tINSERT INTO flight_states (\n\t\t\t\tingestion_run_id,\n\t\t\t\ticao24,\n\t\t\t\tcallsign,\n\t\t\t\tlatitude,\n\t\t\t\tlongitude,\n\t\t\t\tgeometric_altitude_status,\n\t\t\t\tbarometric_altitude_status,\n\t\t\t\tvelocity_mps,\n\t\t\t\theading_degrees,\n\t\t\t\ton_ground,\n\t\t\t\tobserved_at,\n\t\t\t\tposition_source,\n\t\t\t\tsource_name,\n\t\t\t\torigin_country\n\t\t\t)\n\t\t\tVALUES (\n\t\t\t\t$1,\n\t\t\t\t'4B1801',\n\t\t\t\t'AZAL101',\n\t\t\t\t40.4,\n\t\t\t\t49.8,\n\t\t\t\t'unavailable',\n\t\t\t\t'observed',\n\t\t\t\t230,\n\t\t\t\t285,\n\t\t\t\tfalse,\n\t\t\t\t$2,\n\t\t\t\t'mlat',\n\t\t\t\t'adsb.lol',\n\t\t\t\t'Azerbaijan'\n\t\t\t)\n\t\t\`,\n\t\trunID,\n\t\tobservedAt,\n\t)\n\n\titems, err := fixture.repository.GetCurrent(ctx)\n\tif err != nil {\n\t\tt.Fatalf("get current traffic: %v", err)\n\t}\n\tif len(items) != 1 {\n\t\tt.Fatalf("current traffic count = %d, want 1", len(items))\n\t}\n\tif items[0].PositionSource != flightstate.PositionSourceMLAT {\n\t\tt.Fatalf("position source = %q, want %q", items[0].PositionSource, flightstate.PositionSourceMLAT)\n\t}\n\tif items[0].SourceName != "adsb.lol" {\n\t\tt.Fatalf("source name = %q, want adsb.lol", items[0].SourceName)\n\t}\n}\n`,
)

const dto = 'apps/api/internal/http/dto/traffic.go'
replaceExact(
  dto,
  `\tMessageObservedAt  *time.Time                 \`json:"message_observed_at"\`\n\tAircraftModel      string                     \`json:"aircraft_model"\``,
  `\tMessageObservedAt  *time.Time                 \`json:"message_observed_at"\`\n\tPositionSource     *flightstate.PositionSource \`json:"position_source"\`\n\tSourceName         string                     \`json:"source_name"\`\n\tAircraftModel      string                     \`json:"aircraft_model"\``,
)

const handler = 'apps/api/internal/http/handlers/traffic.go'
replaceExact(
  handler,
  `\t\t\tMessageObservedAt:  item.MessageObservedAt,\n\t\t\tAircraftModel:      item.AircraftModel,`,
  `\t\t\tMessageObservedAt:  item.MessageObservedAt,\n\t\t\tPositionSource:     nullablePositionSource(item.PositionSource),\n\t\t\tSourceName:         item.SourceName,\n\t\t\tAircraftModel:      item.AircraftModel,`,
)
replaceExact(
  handler,
  `func toCurrentTrafficItems(items []traffic.CurrentTrafficItem) []dto.CurrentTrafficItem {`,
  `func nullablePositionSource(\n\tvalue flightstate.PositionSource,\n) *flightstate.PositionSource {\n\tif value == flightstate.PositionSourceUnknown {\n\t\treturn nil\n\t}\n\tresult := value\n\treturn &result\n}\n\nfunc toCurrentTrafficItems(items []traffic.CurrentTrafficItem) []dto.CurrentTrafficItem {`,
)
replaceExact(
  handler,
  `\t"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/region"\n`,
  `\t"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"\n\t"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/region"\n`,
)

const verifier = 'scripts/verify-openapi-contract.mjs'
replaceExact(
  verifier,
  `'json:"message_observed_at"'], 'traffic DTO')`,
  `'json:"message_observed_at"','json:"position_source"','json:"source_name"'], 'traffic DTO')`,
)

const specPath = 'openapi/openapi.json'
const spec = JSON.parse(read(specPath))
const currentTraffic = spec.components?.schemas?.CurrentTrafficItem
if (!currentTraffic || currentTraffic.type !== 'object') {
  throw new Error('CurrentTrafficItem schema is unavailable')
}
if (Object.hasOwn(currentTraffic.properties, 'position_source') || Object.hasOwn(currentTraffic.properties, 'source_name')) {
  throw new Error('CurrentTrafficItem provenance fields already exist')
}
currentTraffic.properties.position_source = {
  type: ['string', 'null'],
  enum: ['adsb', 'adsr', 'tisb', 'adsc', 'asterix', 'mlat', 'flarm', null],
  description: 'Canonical position method when observed. Null means the method is unavailable; this is distinct from the data feed identity.',
}
currentTraffic.properties.source_name = {
  type: 'string',
  description: 'Canonical provider/feed identity that supplied the persisted observation. It is not a position-method or accuracy claim.',
}
currentTraffic.required.push('position_source', 'source_name')
const serializedSpec = `${JSON.stringify(spec, null, 2)}\n`
write(specPath, serializedSpec)
write('apps/api/internal/http/apidocs/openapi.json', serializedSpec)

const trafficTypes = 'apps/web/types/traffic.ts'
replaceExact(
  trafficTypes,
  `export type TrafficAltitudeSource =\n  | 'geometric'\n  | 'barometric'\n  | 'ground'\n  | 'none'\n\nexport interface TrafficAircraft {`,
  `export type TrafficAltitudeSource =\n  | 'geometric'\n  | 'barometric'\n  | 'ground'\n  | 'none'\n\nexport type TrafficPositionSource =\n  | 'adsb'\n  | 'adsr'\n  | 'tisb'\n  | 'adsc'\n  | 'asterix'\n  | 'mlat'\n  | 'flarm'\n\nexport interface TrafficAircraft {`,
)
replaceExact(
  trafficTypes,
  `  message_observed_at: string | null\n  aircraft_model: string`,
  `  message_observed_at: string | null\n  position_source?: TrafficPositionSource | null\n  source_name?: string\n  aircraft_model: string`,
)

write(
  'apps/web/lib/traffic/position-provenance.ts',
  `import type {\n  TrafficAircraft,\n  TrafficPositionSource,\n} from '../../types/traffic'\n\nexport type PositionProvenanceStatus = 'observed' | 'unavailable'\n\nexport interface PositionProvenanceEvidence {\n  status: PositionProvenanceStatus\n  method: TrafficPositionSource | null\n  methodLabel: string\n  description: string\n  sourceName: string | null\n}\n\nconst presentations: Record<\n  TrafficPositionSource,\n  { label: string; description: string }\n> = {\n  adsb: {\n    label: 'ADS-B',\n    description: 'Position reported by an ADS-B-equipped emitter.',\n  },\n  adsr: {\n    label: 'ADS-R',\n    description: 'ADS-B position rebroadcast through ADS-R.',\n  },\n  tisb: {\n    label: 'TIS-B',\n    description: 'Traffic-information position rebroadcast through TIS-B.',\n  },\n  adsc: {\n    label: 'ADS-C',\n    description: 'Position reported through aircraft communication data.',\n  },\n  mlat: {\n    label: 'MLAT',\n    description:\n      'Position derived by multilateration; accuracy can vary with receiver geometry.',\n  },\n  asterix: {\n    label: 'ASTERIX',\n    description: 'Position supplied through an ASTERIX surveillance feed.',\n  },\n  flarm: {\n    label: 'FLARM',\n    description: 'Position supplied through a FLARM-compatible feed.',\n  },\n}\n\nexport function buildPositionProvenanceEvidence(\n  aircraft: TrafficAircraft\n): PositionProvenanceEvidence {\n  const method = normalizePositionSource(aircraft.position_source)\n  const sourceName = normalizeSourceName(aircraft.source_name)\n\n  if (method === null) {\n    return {\n      status: 'unavailable',\n      method: null,\n      methodLabel: 'Unavailable',\n      description:\n        'No position-method evidence is available for this observation.',\n      sourceName,\n    }\n  }\n\n  return {\n    status: 'observed',\n    method,\n    methodLabel: presentations[method].label,\n    description: presentations[method].description,\n    sourceName,\n  }\n}\n\nfunction normalizePositionSource(value: unknown): TrafficPositionSource | null {\n  return typeof value === 'string' && Object.hasOwn(presentations, value)\n    ? (value as TrafficPositionSource)\n    : null\n}\n\nfunction normalizeSourceName(value: unknown): string | null {\n  if (typeof value !== 'string') return null\n  const normalized = value.trim()\n  return normalized === '' ? null : normalized\n}\n`,
)

write(
  'apps/web/tests/position-provenance.test.mjs',
  `import assert from 'node:assert/strict'\nimport test from 'node:test'\n\nconst moduleURL = new URL(\n  '../.test-dist/lib/traffic/position-provenance.js',\n  import.meta.url\n)\nconst { buildPositionProvenanceEvidence } = await import(moduleURL.href)\n\nfunction aircraft(overrides = {}) {\n  return {\n    icao24: '4b1801',\n    callsign: 'AZAL101',\n    latitude: 40.4093,\n    longitude: 49.8671,\n    altitude_m: 10668,\n    altitude_status: 'observed',\n    altitude_source: 'barometric',\n    velocity_mps: 230,\n    heading_degrees: 285,\n    on_ground: false,\n    observed_at: '2026-09-09T12:00:00Z',\n    position_observed_at: '2026-09-09T12:00:00Z',\n    message_observed_at: '2026-09-09T12:00:08Z',\n    aircraft_model: 'Airbus A320',\n    airline: 'Azerbaijan Airlines',\n    origin_country: 'Azerbaijan',\n    ...overrides,\n  }\n}\n\ntest('keeps data feed separate from the observed position method', () => {\n  const evidence = buildPositionProvenanceEvidence(\n    aircraft({ position_source: 'mlat', source_name: 'adsb.lol' })\n  )\n  assert.equal(evidence.status, 'observed')\n  assert.equal(evidence.methodLabel, 'MLAT')\n  assert.equal(evidence.sourceName, 'adsb.lol')\n  assert.match(evidence.description, /multilateration/i)\n})\n\ntest('does not infer ADS-B from an ADSB.lol feed name', () => {\n  const evidence = buildPositionProvenanceEvidence(\n    aircraft({ position_source: null, source_name: 'adsb.lol' })\n  )\n  assert.equal(evidence.status, 'unavailable')\n  assert.equal(evidence.method, null)\n  assert.equal(evidence.sourceName, 'adsb.lol')\n})\n\ntest('presents ADS-R, TIS-B and ADS-C as distinct methods', () => {\n  assert.equal(\n    buildPositionProvenanceEvidence(aircraft({ position_source: 'adsr' }))\n      .methodLabel,\n    'ADS-R'\n  )\n  assert.equal(\n    buildPositionProvenanceEvidence(aircraft({ position_source: 'tisb' }))\n      .methodLabel,\n    'TIS-B'\n  )\n  assert.equal(\n    buildPositionProvenanceEvidence(aircraft({ position_source: 'adsc' }))\n      .methodLabel,\n    'ADS-C'\n  )\n})\n\ntest('treats unknown runtime values as unavailable instead of fabricating provenance', () => {\n  const evidence = buildPositionProvenanceEvidence(\n    aircraft({ position_source: 'future-source' })\n  )\n  assert.equal(evidence.status, 'unavailable')\n  assert.equal(evidence.method, null)\n})\n`,
)

const testConfig = 'apps/web/tsconfig.test.json'
replaceExact(
  testConfig,
  `    "lib/traffic/traffic-freshness.ts",\n`,
  `    "lib/traffic/traffic-freshness.ts",\n    "lib/traffic/position-provenance.ts",\n`,
)

const detailPanel = 'apps/web/components/aircraft/aircraft-detail-panel.tsx'
replaceExact(
  detailPanel,
  `import {\n  buildTrafficFreshnessEvidence,\n  formatTrafficEvidenceAge,\n  type TrafficFreshnessEvidence,\n  type TrafficPositionFreshnessStatus,\n} from '@/lib/traffic/traffic-freshness'\n`,
  `import {\n  buildTrafficFreshnessEvidence,\n  formatTrafficEvidenceAge,\n  type TrafficFreshnessEvidence,\n  type TrafficPositionFreshnessStatus,\n} from '@/lib/traffic/traffic-freshness'\nimport {\n  buildPositionProvenanceEvidence,\n  type PositionProvenanceEvidence,\n} from '@/lib/traffic/position-provenance'\n`,
)
replaceExact(
  detailPanel,
  `  const freshnessEvidence = aircraft\n    ? buildTrafficFreshnessEvidence(aircraft, trafficSnapshotUpdatedAt)\n    : null\n\n  return (`,
  `  const freshnessEvidence = aircraft\n    ? buildTrafficFreshnessEvidence(aircraft, trafficSnapshotUpdatedAt)\n    : null\n  const positionProvenanceEvidence = aircraft\n    ? buildPositionProvenanceEvidence(aircraft)\n    : null\n\n  return (`,
)
replaceExact(
  detailPanel,
  `        {freshnessEvidence ? (\n          <ObservationFreshnessSection evidence={freshnessEvidence} />\n        ) : null}\n\n        <RouteContextSection`,
  `        {freshnessEvidence ? (\n          <ObservationFreshnessSection evidence={freshnessEvidence} />\n        ) : null}\n\n        {positionProvenanceEvidence ? (\n          <PositionProvenanceSection evidence={positionProvenanceEvidence} />\n        ) : null}\n\n        <RouteContextSection`,
)
replaceExact(
  detailPanel,
  `function ObservationFreshnessSection({`,
  `function PositionProvenanceSection({\n  evidence,\n}: {\n  evidence: PositionProvenanceEvidence\n}) {\n  return (\n    <section\n      className='mt-4 border-t border-white/10 pt-4'\n      aria-labelledby='position-provenance-title'\n    >\n      <SectionHeading\n        id='position-provenance-title'\n        label='Position provenance'\n        evidence='Observed metadata'\n      />\n      <dl className='mt-2 grid grid-cols-2 gap-px overflow-hidden rounded-lg border border-white/10 bg-white/10'>\n        <div className='min-w-0 bg-[#202328] p-2.5'>\n          <dt className='text-[9px] font-semibold uppercase tracking-[0.12em] text-slate-500'>\n            Method\n          </dt>\n          <dd className='mt-1 text-xs font-semibold text-slate-100'>\n            {evidence.methodLabel}\n          </dd>\n        </div>\n        <div className='min-w-0 bg-[#202328] p-2.5'>\n          <dt className='text-[9px] font-semibold uppercase tracking-[0.12em] text-slate-500'>\n            Data feed\n          </dt>\n          <dd className='mt-1 break-words text-xs font-semibold text-slate-100'>\n            {evidence.sourceName ?? 'Unavailable'}\n          </dd>\n        </div>\n      </dl>\n      <p className='mt-2 text-[10px] leading-4 text-slate-500'>\n        {evidence.description}\n      </p>\n    </section>\n  )\n}\n\nfunction ObservationFreshnessSection({`,
)

const mockAPI = 'apps/web/e2e/mock-api.mjs'
replaceExact(
  mockAPI,
  `    message_observed_at: '2026-08-04T18:00:03Z',\n    aircraft_model: 'Airbus A320',`,
  `    message_observed_at: '2026-08-04T18:00:03Z',\n    position_source: 'mlat',\n    source_name: 'adsb.lol',\n    aircraft_model: 'Airbus A320',`,
)
replaceExact(
  mockAPI,
  `    message_observed_at: '2026-08-04T18:00:05.500Z',\n    aircraft_model: 'Boeing 737-800',`,
  `    message_observed_at: '2026-08-04T18:00:05.500Z',\n    position_source: 'adsb',\n    source_name: 'adsb.lol',\n    aircraft_model: 'Boeing 737-800',`,
)

write(
  'docs/217_READSB_POSITION_SOURCE_PROVENANCE.md',
  `# Document 217 — readsb-Inspired Position Source Provenance\n\n## Status\n\nR2 engineering implementation on a stacked feature branch. This document does not authorize merge.\n\n## Goal\n\nPreserve the position-method evidence already present in readsb-compatible provider payloads and expose it separately from the provider/feed identity.\n\nThe project must never infer a position method from a provider brand. An observation supplied by ADSB.lol may be ADS-B, MLAT, ADS-R, TIS-B, ADS-C, or unavailable.\n\n## Canonical mapping\n\n| readsb type | GFA position_source |\n| --- | --- |\n| adsb_icao, adsb_icao_nt, adsb_other | adsb |\n| adsr_icao, adsr_other | adsr |\n| tisb_icao, tisb_other, tisb_trackfile | tisb |\n| adsc | adsc |\n| mlat | mlat |\n| mode_s, other, empty, unknown future values | unavailable |\n\nMode-S is not promoted to a position method because the readsb contract describes it as transponder data without a transmitted position.\n\n## Evidence separation\n\n\`\`\`text\nsource_name      = provider/feed identity\nposition_source  = observed position method\n\`\`\`\n\nExamples:\n\n\`\`\`text\nsource_name=adsb.lol\nposition_source=mlat\n\`\`\`\n\nis valid and must not be rewritten to ADS-B.\n\n## Persistence\n\nMigration 031 expands the existing flight_states.position_source constraint with adsr, tisb, and adsc. Existing adsb, asterix, mlat, flarm, and unavailable values remain valid. No historical row is rewritten.\n\n## API\n\nCurrent Traffic exposes:\n\n- position_source: nullable canonical method;\n- source_name: provider/feed identity.\n\nNull position_source means evidence is unavailable. It is not a negative quality judgment.\n\n## Frontend\n\nAircraft Detail displays a Position provenance section with:\n\n- Method;\n- Data feed;\n- a method-specific factual explanation.\n\nNo synthetic confidence score, accuracy percentage, provider ranking, or preferred-source claim is introduced.\n\n## Cost boundary\n\n- new provider: NO\n- new external call: NO\n- new service: NO\n- Redis/Valkey: NO\n- Python runtime: NO\n- paid infrastructure: NO\n- hardware requirement: NO\n- additional monetary cost: 0 RUB\n\n## Stack dependency\n\nR2 is based on R1 Position Freshness and therefore cannot become a canonical main merge candidate until PR #174 and R1 are merged and R2 is reconciled onto the resulting main. Any reconciliation changes the exact head and requires fresh validation and fresh merge authorization.\n`,
)

const index = 'docs/DOCUMENT_INDEX.md'
replaceExact(
  index,
  `## Document 216 — readsb-Inspired Position Freshness\n\n\`216_READSB_POSITION_FRESHNESS.md\` records the zero-cost R1 separation of position observation time from latest message/contact time, readsb \`seen_pos\` semantics, OpenSky \`time_position\` versus \`last_contact\`, nullable persistence, public traffic contract evolution, frontend freshness evidence, and the rule that message freshness never upgrades stale position evidence.`,
  `## Document 216 — readsb-Inspired Position Freshness\n\n\`216_READSB_POSITION_FRESHNESS.md\` records the zero-cost R1 separation of position observation time from latest message/contact time, readsb \`seen_pos\` semantics, OpenSky \`time_position\` versus \`last_contact\`, nullable persistence, public traffic contract evolution, frontend freshness evidence, and the rule that message freshness never upgrades stale position evidence.\n\n## Document 217 — readsb-Inspired Position Source Provenance\n\n\`217_READSB_POSITION_SOURCE_PROVENANCE.md\` records the zero-cost R2 preservation of readsb position-method evidence, strict separation of provider/feed identity from position method, canonical ADS-B/ADS-R/TIS-B/ADS-C/MLAT mapping, nullable public API semantics, and Aircraft Detail provenance presentation without synthetic confidence or accuracy claims.`,
)

execFileSync('node', ['scripts/generate-openapi-client.mjs', '--write'], {
  stdio: 'inherit',
})
