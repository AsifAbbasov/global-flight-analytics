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

const trafficModel = 'apps/api/internal/domain/traffic/model.go'
replaceExact(
  trafficModel,
  `\tVelocityMPS       float64\n\tHeadingDegrees    float64\n\tOnGround          bool\n`,
  `\tVelocityMPS       float64\n\tHeadingDegrees    float64\n\tVerticalRateMPS   *float64\n\tOnGround          bool\n`,
)

const trafficRepository = 'apps/api/internal/repository/postgres/traffic_repository.go'
replaceExact(
  trafficRepository,
  `import (\n\t"context"\n`,
  `import (\n\t"context"\n\t"math"\n`,
)
replaceExact(
  trafficRepository,
  `\t\t\tfs.velocity_mps,\n\t\t\tfs.heading_degrees,\n\t\t\tfs.on_ground,`,
  `\t\t\tfs.velocity_mps,\n\t\t\tfs.heading_degrees,\n\t\t\tfs.vertical_rate_mps,\n\t\t\tfs.on_ground,`,
  2,
)
replaceExact(
  trafficRepository,
  `\t\tvar messageObservedAt pgtype.Timestamptz\n\t\tvar positionSource string\n`,
  `\t\tvar messageObservedAt pgtype.Timestamptz\n\t\tvar positionSource string\n\t\tvar verticalRate pgtype.Float8\n`,
)
replaceExact(
  trafficRepository,
  `\t\t\t&item.VelocityMPS,\n\t\t\t&item.HeadingDegrees,\n\t\t\t&item.OnGround,`,
  `\t\t\t&item.VelocityMPS,\n\t\t\t&item.HeadingDegrees,\n\t\t\t&verticalRate,\n\t\t\t&item.OnGround,`,
)
replaceExact(
  trafficRepository,
  `\t\tif messageObservedAt.Valid {\n`,
  `\t\titem.VerticalRateMPS = nullableTrafficVerticalRate(verticalRate)\n\n\t\tif messageObservedAt.Valid {\n`,
)
replaceExact(
  trafficRepository,
  `func nullableTrafficAltitude(\n`,
  `func nullableTrafficVerticalRate(\n\tvalue pgtype.Float8,\n) *float64 {\n\tif !value.Valid || math.IsNaN(value.Float64) || math.IsInf(value.Float64, 0) {\n\t\treturn nil\n\t}\n\n\tresult := value.Float64\n\treturn &result\n}\n\nfunc nullableTrafficAltitude(\n`,
)

const trafficDTO = 'apps/api/internal/http/dto/traffic.go'
replaceExact(
  trafficDTO,
  `\tVelocityMPS        float64                     \`json:"velocity_mps"\`\n\tHeadingDegrees     float64                     \`json:"heading_degrees"\`\n\tOnGround           bool                        \`json:"on_ground"\`\n`,
  `\tVelocityMPS        float64                     \`json:"velocity_mps"\`\n\tHeadingDegrees     float64                     \`json:"heading_degrees"\`\n\tVerticalRateMPS    *float64                    \`json:"vertical_rate_mps"\`\n\tOnGround           bool                        \`json:"on_ground"\`\n`,
)

const trafficHandler = 'apps/api/internal/http/handlers/traffic.go'
replaceExact(
  trafficHandler,
  `\t\t\tVelocityMPS:        item.VelocityMPS,\n\t\t\tHeadingDegrees:     item.HeadingDegrees,\n\t\t\tOnGround:           item.OnGround,\n`,
  `\t\t\tVelocityMPS:        item.VelocityMPS,\n\t\t\tHeadingDegrees:     item.HeadingDegrees,\n\t\t\tVerticalRateMPS:    item.VerticalRateMPS,\n\t\t\tOnGround:           item.OnGround,\n`,
)

write(
  'apps/api/internal/http/handlers/traffic_vertical_rate_test.go',
  `package handlers\n\nimport (\n\t"testing"\n\n\t"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/traffic"\n)\n\nfunc TestToCurrentTrafficItemsPreservesVerticalRateEvidence(t *testing.T) {\n\tclimb := 4.25\n\tdescent := -3.5\n\n\tresult := toCurrentTrafficItems([]traffic.CurrentTrafficItem{\n\t\t{ICAO24: "climb01", VerticalRateMPS: &climb},\n\t\t{ICAO24: "desc01", VerticalRateMPS: &descent},\n\t\t{ICAO24: "none01", VerticalRateMPS: nil},\n\t})\n\n\tif len(result) != 3 {\n\t\tt.Fatalf("traffic DTO count = %d, want 3", len(result))\n\t}\n\tif result[0].VerticalRateMPS == nil || *result[0].VerticalRateMPS != climb {\n\t\tt.Fatalf("climb vertical rate = %#v, want %v", result[0].VerticalRateMPS, climb)\n\t}\n\tif result[1].VerticalRateMPS == nil || *result[1].VerticalRateMPS != descent {\n\t\tt.Fatalf("descent vertical rate = %#v, want %v", result[1].VerticalRateMPS, descent)\n\t}\n\tif result[2].VerticalRateMPS != nil {\n\t\tt.Fatalf("unavailable vertical rate became numeric: %v", *result[2].VerticalRateMPS)\n\t}\n}\n`,
)

const trafficIntegration = 'apps/api/internal/repository/postgres/traffic_altitude_semantics_integration_test.go'
const existingRunInsert = `\tmustExecTrafficAltitudeSQL(\n\t\tt,\n\t\tfixture.pool,\n\t\t\`\n\t\t\tINSERT INTO ingestion_runs (\n\t\t\t\tid,\n\t\t\t\tfinished_at,\n\t\t\t\tstatus,\n\t\t\t\tcreated_at\n\t\t\t)\n\t\t\tVALUES ($1, $2, 'success', $2)\n\t\t\`,\n\t\trunID,\n\t\tfinishedAt,\n\t)\n`
replaceExact(
  trafficIntegration,
  existingRunInsert,
  `\tinsertTrafficTestRun(t, fixture.pool, runID, finishedAt)\n`,
)
replaceExact(
  trafficIntegration,
  `\t\t\t\tvelocity_mps double precision,\n\t\t\t\theading_degrees double precision,\n\t\t\t\ton_ground boolean,\n`,
  `\t\t\t\tvelocity_mps double precision,\n\t\t\t\theading_degrees double precision,\n\t\t\t\tvertical_rate_mps double precision,\n\t\t\t\ton_ground boolean,\n`,
)
replaceExact(
  trafficIntegration,
  `func insertTrafficAltitudeState(\n`,
  `func insertTrafficTestRun(\n\tt *testing.T,\n\tpool *pgxpool.Pool,\n\trunID string,\n\tfinishedAt time.Time,\n) {\n\tt.Helper()\n\n\tmustExecTrafficAltitudeSQL(\n\t\tt,\n\t\tpool,\n\t\t\`\n\t\t\tINSERT INTO ingestion_runs (\n\t\t\t\tid,\n\t\t\t\tfinished_at,\n\t\t\t\tstatus,\n\t\t\t\tcreated_at\n\t\t\t)\n\t\t\tVALUES ($1, $2, 'success', $2)\n\t\t\`,\n\t\trunID,\n\t\tfinishedAt,\n\t)\n}\n\nfunc insertTrafficAltitudeState(\n`,
)
replaceExact(
  trafficIntegration,
  `type trafficAltitudeFixture struct {\n`,
  `func TestTrafficRepositoryPreservesVerticalRateAvailability(t *testing.T) {\n\tfixture := newTrafficAltitudeFixture(t)\n\tctx := context.Background()\n\trunID := "22222222-2222-2222-2222-222222222222"\n\tfinishedAt := time.Date(2026, time.September, 9, 15, 0, 0, 0, time.UTC)\n\n\tinsertTrafficTestRun(t, fixture.pool, runID, finishedAt)\n\tinsertTrafficAltitudeState(t, fixture.pool, runID, "VRATE01", 40.0, 49.0, intPointer(1500), "observed", intPointer(1450), "observed", false, finishedAt)\n\tinsertTrafficAltitudeState(t, fixture.pool, runID, "VRATE02", 40.1, 49.1, intPointer(1400), "observed", intPointer(1350), "observed", false, finishedAt.Add(time.Second))\n\tinsertTrafficAltitudeState(t, fixture.pool, runID, "VRATE03", 40.2, 49.2, intPointer(1300), "observed", intPointer(1250), "observed", false, finishedAt.Add(2*time.Second))\n\n\tmustExecTrafficAltitudeSQL(\n\t\tt,\n\t\tfixture.pool,\n\t\t\`\n\t\t\tUPDATE flight_states\n\t\t\tSET vertical_rate_mps = CASE icao24\n\t\t\t\tWHEN 'VRATE01' THEN 4.25\n\t\t\t\tWHEN 'VRATE02' THEN -3.5\n\t\t\t\tELSE NULL\n\t\t\tEND\n\t\t\tWHERE ingestion_run_id = $1\n\t\t\`,\n\t\trunID,\n\t)\n\n\titems, err := fixture.repository.GetCurrent(ctx)\n\tif err != nil {\n\t\tt.Fatalf("get current traffic: %v", err)\n\t}\n\tif len(items) != 3 {\n\t\tt.Fatalf("current traffic count = %d, want 3", len(items))\n\t}\n\n\tassertVerticalRate := func(index int, icao24 string, expected *float64) {\n\t\tt.Helper()\n\t\titem := items[index]\n\t\tif item.ICAO24 != icao24 {\n\t\t\tt.Fatalf("item %d icao24 = %s, want %s", index, item.ICAO24, icao24)\n\t\t}\n\t\tif expected == nil {\n\t\t\tif item.VerticalRateMPS != nil {\n\t\t\t\tt.Fatalf("%s vertical rate = %v, want unavailable", icao24, *item.VerticalRateMPS)\n\t\t\t}\n\t\t\treturn\n\t\t}\n\t\tif item.VerticalRateMPS == nil || *item.VerticalRateMPS != *expected {\n\t\t\tt.Fatalf("%s vertical rate = %#v, want %v", icao24, item.VerticalRateMPS, *expected)\n\t\t}\n\t}\n\n\tclimb := 4.25\n\tdescent := -3.5\n\tassertVerticalRate(0, "VRATE01", &climb)\n\tassertVerticalRate(1, "VRATE02", &descent)\n\tassertVerticalRate(2, "VRATE03", nil)\n}\n\ntype trafficAltitudeFixture struct {\n`,
)

function updateCurrentTrafficSchema(path) {
  const spec = JSON.parse(read(path))
  const schema = spec?.components?.schemas?.CurrentTrafficItem
  if (!schema || typeof schema !== 'object' || !schema.properties) {
    throw new Error(`${path}: CurrentTrafficItem schema missing`)
  }
  if (Object.hasOwn(schema.properties, 'vertical_rate_mps')) {
    throw new Error(`${path}: vertical_rate_mps already exists`)
  }

  const properties = {}
  for (const [name, value] of Object.entries(schema.properties)) {
    properties[name] = value
    if (name === 'heading_degrees') {
      properties.vertical_rate_mps = {
        type: ['number', 'null'],
        description:
          'Observed vertical rate in metres per second. Positive values indicate climb, negative values indicate descent, and null means the provider did not supply usable vertical-rate evidence.',
      }
    }
  }
  schema.properties = properties

  if (!Array.isArray(schema.required)) {
    throw new Error(`${path}: CurrentTrafficItem required list missing`)
  }
  if (schema.required.includes('vertical_rate_mps')) {
    throw new Error(`${path}: vertical_rate_mps already required`)
  }
  const headingIndex = schema.required.indexOf('heading_degrees')
  if (headingIndex < 0) {
    throw new Error(`${path}: heading_degrees required field missing`)
  }
  schema.required.splice(headingIndex + 1, 0, 'vertical_rate_mps')
  write(path, `${JSON.stringify(spec, null, 2)}\n`)
}

updateCurrentTrafficSchema('openapi/openapi.json')
updateCurrentTrafficSchema('apps/api/internal/http/apidocs/openapi.json')

const openAPIVerifier = 'scripts/verify-openapi-contract.mjs'
replaceExact(
  openAPIVerifier,
  `'json:"altitude_source"','json:"observed_at"'`,
  `'json:"altitude_source"','json:"vertical_rate_mps"','json:"observed_at"'`,
)

const webTrafficTypes = 'apps/web/types/traffic.ts'
replaceExact(
  webTrafficTypes,
  `  velocity_mps: number\n  heading_degrees: number\n  on_ground: boolean\n`,
  `  velocity_mps: number\n  heading_degrees: number\n  vertical_rate_mps: number | null\n  on_ground: boolean\n`,
)

const mockAPI = 'apps/web/e2e/mock-api.mjs'
replaceExact(
  mockAPI,
  `    velocity_mps: 230,\n    heading_degrees: 285,\n    on_ground: false,\n`,
  `    velocity_mps: 230,\n    heading_degrees: 285,\n    vertical_rate_mps: 6.2,\n    on_ground: false,\n`,
)
replaceExact(
  mockAPI,
  `    velocity_mps: 218,\n    heading_degrees: 92,\n    on_ground: false,\n`,
  `    velocity_mps: 218,\n    heading_degrees: 92,\n    vertical_rate_mps: -4.8,\n    on_ground: false,\n`,
)

write(
  'apps/web/lib/traffic/vertical-motion.ts',
  `import type { TrafficAircraft } from '../../types/traffic'\n\nexport type VerticalMotionStatus =\n  | 'climbing'\n  | 'descending'\n  | 'level'\n  | 'ground'\n  | 'unavailable'\n\nexport const defaultVerticalMotionDeadbandMPS = 0.5\n\nexport interface VerticalMotionEvidence {\n  status: VerticalMotionStatus\n  label: string\n  verticalRateMPS: number | null\n  verticalRateFeetPerMinute: number | null\n  displayRate: string\n  description: string\n}\n\nexport function buildVerticalMotionEvidence(\n  aircraft: TrafficAircraft,\n  deadbandMPS = defaultVerticalMotionDeadbandMPS\n): VerticalMotionEvidence {\n  const verticalRateMPS = finiteNumberOrNull(aircraft.vertical_rate_mps)\n  const validDeadband =\n    Number.isFinite(deadbandMPS) && deadbandMPS >= 0\n      ? deadbandMPS\n      : defaultVerticalMotionDeadbandMPS\n\n  if (aircraft.on_ground) {\n    return evidence(\n      'ground',\n      'On ground',\n      verticalRateMPS,\n      'Ground state is observed; vertical-rate evidence is shown separately when available.'\n    )\n  }\n\n  if (verticalRateMPS === null) {\n    return evidence(\n      'unavailable',\n      'Unavailable',\n      null,\n      'No usable vertical-rate evidence is available for this observation.'\n    )\n  }\n\n  if (verticalRateMPS > validDeadband) {\n    return evidence(\n      'climbing',\n      '↑ Climbing',\n      verticalRateMPS,\n      'Positive observed vertical rate exceeds the presentation deadband.'\n    )\n  }\n\n  if (verticalRateMPS < -validDeadband) {\n    return evidence(\n      'descending',\n      '↓ Descending',\n      verticalRateMPS,\n      'Negative observed vertical rate exceeds the presentation deadband.'\n    )\n  }\n\n  return evidence(\n    'level',\n    'Level',\n    verticalRateMPS,\n    `Observed vertical rate is within the ±${validDeadband.toFixed(1)} m/s presentation deadband.`\n  )\n}\n\nfunction evidence(\n  status: VerticalMotionStatus,\n  label: string,\n  verticalRateMPS: number | null,\n  description: string\n): VerticalMotionEvidence {\n  const verticalRateFeetPerMinute =\n    verticalRateMPS === null ? null : verticalRateMPS * 196.8503937007874\n\n  return {\n    status,\n    label,\n    verticalRateMPS,\n    verticalRateFeetPerMinute,\n    displayRate: formatVerticalRate(verticalRateMPS, verticalRateFeetPerMinute),\n    description,\n  }\n}\n\nexport function formatVerticalRate(\n  verticalRateMPS: number | null,\n  verticalRateFeetPerMinute =\n    verticalRateMPS === null ? null : verticalRateMPS * 196.8503937007874\n): string {\n  if (verticalRateMPS === null || verticalRateFeetPerMinute === null) {\n    return 'Unavailable'\n  }\n\n  const mpsPrefix = verticalRateMPS > 0 ? '+' : ''\n  const fpmRounded = Math.round(verticalRateFeetPerMinute)\n  const fpmPrefix = fpmRounded > 0 ? '+' : ''\n\n  return `${mpsPrefix}${verticalRateMPS.toFixed(1)} m/s (${fpmPrefix}${fpmRounded.toLocaleString('en-US')} ft/min)`\n}\n\nfunction finiteNumberOrNull(value: unknown): number | null {\n  return typeof value === 'number' && Number.isFinite(value) ? value : null\n}\n`,
)

write(
  'apps/web/tests/vertical-motion.test.mjs',
  `import assert from 'node:assert/strict'\nimport test from 'node:test'\n\nconst moduleURL = new URL('../.test-dist/lib/traffic/vertical-motion.js', import.meta.url)\nconst {\n  buildVerticalMotionEvidence,\n  defaultVerticalMotionDeadbandMPS,\n  formatVerticalRate,\n} = await import(moduleURL.href)\n\nfunction aircraft(overrides = {}) {\n  return {\n    icao24: '4b1801',\n    callsign: 'AZAL101',\n    latitude: 40.4093,\n    longitude: 49.8671,\n    altitude_m: 10668,\n    altitude_status: 'observed',\n    altitude_source: 'barometric',\n    velocity_mps: 230,\n    heading_degrees: 285,\n    vertical_rate_mps: 0,\n    on_ground: false,\n    observed_at: '2026-09-09T12:00:00Z',\n    position_observed_at: '2026-09-09T12:00:00Z',\n    message_observed_at: '2026-09-09T12:00:03Z',\n    position_source: 'adsb',\n    source_name: 'adsb.lol',\n    aircraft_model: 'Airbus A320',\n    airline: 'Azerbaijan Airlines',\n    origin_country: 'Azerbaijan',\n    ...overrides,\n  }\n}\n\ntest('classifies positive observed vertical rate as climbing', () => {\n  const evidence = buildVerticalMotionEvidence(aircraft({ vertical_rate_mps: 4.25 }))\n  assert.equal(evidence.status, 'climbing')\n  assert.equal(evidence.label, '↑ Climbing')\n  assert.equal(evidence.verticalRateMPS, 4.25)\n  assert.match(evidence.displayRate, /^\\+4\\.3 m\\/s/)\n})\n\ntest('classifies negative observed vertical rate as descending', () => {\n  const evidence = buildVerticalMotionEvidence(aircraft({ vertical_rate_mps: -3.5 }))\n  assert.equal(evidence.status, 'descending')\n  assert.equal(evidence.label, '↓ Descending')\n  assert.match(evidence.displayRate, /^-3\\.5 m\\/s/)\n})\n\ntest('uses a bounded near-level deadband', () => {\n  assert.equal(defaultVerticalMotionDeadbandMPS, 0.5)\n  assert.equal(buildVerticalMotionEvidence(aircraft({ vertical_rate_mps: 0.5 })).status, 'level')\n  assert.equal(buildVerticalMotionEvidence(aircraft({ vertical_rate_mps: -0.5 })).status, 'level')\n})\n\ntest('keeps unavailable vertical-rate evidence unavailable', () => {\n  const evidence = buildVerticalMotionEvidence(aircraft({ vertical_rate_mps: null }))\n  assert.equal(evidence.status, 'unavailable')\n  assert.equal(evidence.verticalRateMPS, null)\n  assert.equal(evidence.displayRate, 'Unavailable')\n})\n\ntest('ground state takes precedence over a motion classification', () => {\n  const evidence = buildVerticalMotionEvidence(\n    aircraft({ on_ground: true, vertical_rate_mps: 2 })\n  )\n  assert.equal(evidence.status, 'ground')\n  assert.equal(evidence.label, 'On ground')\n})\n\ntest('formats feet per minute without changing canonical metres-per-second evidence', () => {\n  assert.equal(formatVerticalRate(1), '+1.0 m/s (+197 ft/min)')\n  assert.equal(formatVerticalRate(0), '0.0 m/s (0 ft/min)')\n})\n`,
)

const webTestConfig = 'apps/web/tsconfig.test.json'
replaceExact(
  webTestConfig,
  `    "lib/traffic/traffic-freshness.ts",\n`,
  `    "lib/traffic/traffic-freshness.ts",\n    "lib/traffic/vertical-motion.ts",\n`,
)

const aircraftPanel = 'apps/web/components/aircraft/aircraft-detail-panel.tsx'
replaceExact(
  aircraftPanel,
  `import {\n  buildPositionProvenanceEvidence,\n  type PositionProvenanceEvidence,\n} from '@/lib/traffic/position-provenance'\n`,
  `import {\n  buildPositionProvenanceEvidence,\n  type PositionProvenanceEvidence,\n} from '@/lib/traffic/position-provenance'\nimport {\n  buildVerticalMotionEvidence,\n  type VerticalMotionEvidence,\n  type VerticalMotionStatus,\n} from '@/lib/traffic/vertical-motion'\n`,
)
replaceExact(
  aircraftPanel,
  `  const positionProvenanceEvidence = aircraft\n    ? buildPositionProvenanceEvidence(aircraft)\n    : null\n`,
  `  const positionProvenanceEvidence = aircraft\n    ? buildPositionProvenanceEvidence(aircraft)\n    : null\n  const verticalMotionEvidence = aircraft\n    ? buildVerticalMotionEvidence(aircraft)\n    : null\n`,
)
replaceExact(
  aircraftPanel,
  `        {freshnessEvidence ? (\n          <ObservationFreshnessSection evidence={freshnessEvidence} />\n        ) : null}\n`,
  `        {verticalMotionEvidence ? (\n          <VerticalMotionSection evidence={verticalMotionEvidence} />\n        ) : null}\n\n        {freshnessEvidence ? (\n          <ObservationFreshnessSection evidence={freshnessEvidence} />\n        ) : null}\n`,
)
replaceExact(
  aircraftPanel,
  `function SectionHeading({\n`,
  `function VerticalMotionSection({\n  evidence,\n}: {\n  evidence: VerticalMotionEvidence\n}) {\n  return (\n    <section\n      className='mt-4 border-t border-white/10 pt-4'\n      aria-labelledby='vertical-motion-title'\n    >\n      <div className='flex flex-wrap items-start justify-between gap-3'>\n        <div>\n          <h4 id='vertical-motion-title' className='text-xs font-semibold text-slate-100'>\n            Vertical motion\n          </h4>\n          <p className='mt-0.5 text-[10px] leading-4 text-slate-500'>\n            Provider-observed vertical rate; movement label is presentation-only.\n          </p>\n        </div>\n        <VerticalMotionBadge status={evidence.status} label={evidence.label} />\n      </div>\n\n      <dl className='mt-2 grid gap-px overflow-hidden rounded-lg border border-white/10 bg-white/10 sm:grid-cols-2'>\n        <div className='bg-[#202328] p-2.5'>\n          <dt className='text-[9px] font-semibold uppercase tracking-[0.12em] text-slate-500'>\n            Vertical rate\n          </dt>\n          <dd className='mt-1 text-xs font-semibold text-slate-100'>\n            {evidence.displayRate}\n          </dd>\n        </div>\n        <div className='bg-[#202328] p-2.5'>\n          <dt className='text-[9px] font-semibold uppercase tracking-[0.12em] text-slate-500'>\n            Evidence note\n          </dt>\n          <dd className='mt-1 text-[10px] leading-4 text-slate-400'>\n            {evidence.description}\n          </dd>\n        </div>\n      </dl>\n    </section>\n  )\n}\n\nfunction VerticalMotionBadge({\n  status,\n  label,\n}: {\n  status: VerticalMotionStatus\n  label: string\n}) {\n  const className: Record<VerticalMotionStatus, string> = {\n    climbing: 'border-emerald-300/30 bg-emerald-300/10 text-emerald-200',\n    descending: 'border-sky-300/30 bg-sky-300/10 text-sky-200',\n    level: 'border-slate-300/20 bg-slate-300/10 text-slate-200',\n    ground: 'border-amber-300/30 bg-amber-300/10 text-amber-200',\n    unavailable: 'border-slate-600 bg-slate-800/80 text-slate-400',\n  }\n\n  return (\n    <span\n      className={\`rounded-full border px-2 py-0.5 text-[10px] font-bold tracking-[0.08em] ${className[status]}\`}\n    >\n      {label}\n    </span>\n  )\n}\n\nfunction SectionHeading({\n`,
)

write(
  'docs/218_LIVE_VERTICAL_RATE_AND_MOTION.md',
  `# Document 218 — Live Vertical Rate and Motion\n\nStatus: R3 engineering implementation complete; canonical stacked-PR validation pending.\n\n## Purpose\n\nR3 surfaces the vertical-rate evidence that GFA already persists for flight states and turns it into a bounded live presentation for the Aircraft Detail experience. The feature does not add a new provider, polling path, database table, paid service or background worker.\n\n## Evidence source\n\nThe canonical stored value is \`flight_states.vertical_rate_mps\`. Existing provider adapters already normalize usable vertical-rate telemetry into metres per second before persistence:\n\n- OpenSky supplies vertical rate directly in metres per second when available.\n- readsb-compatible providers normalize barometric rate from feet per minute into metres per second.\n\nCurrent Traffic now exposes that persisted value as nullable \`vertical_rate_mps\`. Null means no usable vertical-rate evidence is available for that observation.\n\n## Frontend interpretation\n\nThe frontend preserves the numeric value and applies a presentation-only deadband of ±0.5 m/s:\n\n- greater than +0.5 m/s → \`↑ Climbing\`\n- less than -0.5 m/s → \`↓ Descending\`\n- within the inclusive ±0.5 m/s band → \`Level\`\n- observed \`on_ground=true\` → \`On ground\` takes presentation precedence\n- missing/non-finite evidence → \`Unavailable\`\n\nThe UI also shows an approximate feet-per-minute conversion for aviation readability while metres per second remains the canonical API evidence.\n\n## Claim boundary\n\nThe movement label is not a flight-phase detector and does not claim pilot intent, cleared altitude, autopilot mode, climb/descent target, turbulence state or future motion. A single observed vertical-rate sample can fluctuate and must not be interpreted as a complete trajectory trend.\n\n## API contract\n\n\`CurrentTrafficItem.vertical_rate_mps\` is required as a response property but nullable. This makes absence explicit rather than silently coercing missing evidence to zero.\n\n## Persistence and cost\n\nNo migration is required because \`vertical_rate_mps\` already exists in the canonical flight-state schema. R3 only connects existing persisted evidence to Current Traffic and the frontend. Infrastructure cost remains zero.\n\n## Validation boundary\n\nBranch-local engineering validation covers Go tests/vet, Stage14 strict audit, OpenAPI contract and generated-client drift, dependency policy checks, API client checks, frontend lint/typecheck/tests and production build. Canonical CI, Vercel, review and merge evidence remain separate stacked-PR gates.\n`,
)

const documentIndex = 'docs/DOCUMENT_INDEX.md'
const indexContent = read(documentIndex)
if (indexContent.includes('218_LIVE_VERTICAL_RATE_AND_MOTION.md')) {
  throw new Error('Document 218 already registered')
}
write(
  documentIndex,
  `${indexContent.trimEnd()}\n\n## Document 218 — Live Vertical Rate and Motion\n\n\`218_LIVE_VERTICAL_RATE_AND_MOTION.md\` records the zero-cost R3 exposure of persisted vertical-rate evidence through Current Traffic, the bounded climbing/descending/level presentation deadband, explicit nullable-evidence behavior and the non-flight-phase claim boundary.\n`,
)

execFileSync('node', ['scripts/generate-openapi-client.mjs', '--write'], {
  stdio: 'inherit',
})
