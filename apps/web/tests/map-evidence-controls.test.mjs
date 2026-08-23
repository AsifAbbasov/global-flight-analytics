import assert from 'node:assert/strict'
import test from 'node:test'

const moduleURL = new URL(
  '../.test-dist/lib/map/map-evidence-controls.js',
  import.meta.url
)
const importedModule = await import(moduleURL.href)
const controls = importedModule.default ?? importedModule
const {
  defaultMapEvidenceVisibility,
  toggleTrajectoryVisibility,
  toggleProjectionVisibility,
  shouldRenderTrajectory,
  shouldRenderProjection,
} = controls

test('map evidence is visible by default and toggles independently', () => {
  assert.deepEqual(defaultMapEvidenceVisibility, {
    trajectory: true,
    projection: true,
  })

  const trajectoryHidden = toggleTrajectoryVisibility(
    defaultMapEvidenceVisibility
  )
  assert.deepEqual(trajectoryHidden, {
    trajectory: false,
    projection: true,
  })

  const projectionHidden = toggleProjectionVisibility(
    defaultMapEvidenceVisibility
  )
  assert.deepEqual(projectionHidden, {
    trajectory: true,
    projection: false,
  })
})

test('trajectory visibility requires both enabled state and real features', () => {
  assert.equal(shouldRenderTrajectory(defaultMapEvidenceVisibility, 2), true)
  assert.equal(shouldRenderTrajectory(defaultMapEvidenceVisibility, 0), false)
  assert.equal(shouldRenderTrajectory({ trajectory: false, projection: true }, 2), false)
  assert.equal(shouldRenderTrajectory(defaultMapEvidenceVisibility, Number.NaN), false)
})

test('projection visibility requires both enabled state and real points', () => {
  assert.equal(shouldRenderProjection(defaultMapEvidenceVisibility, 3), true)
  assert.equal(shouldRenderProjection(defaultMapEvidenceVisibility, 0), false)
  assert.equal(shouldRenderProjection({ trajectory: true, projection: false }, 3), false)
  assert.equal(shouldRenderProjection(defaultMapEvidenceVisibility, -1), false)
})
