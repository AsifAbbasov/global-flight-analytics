'use client'

import { useEffect, useRef } from 'react'
import * as THREE from 'three'

import type { TrafficAircraft } from '@/types/traffic'

interface TrafficGlobeProps {
  aircraft: TrafficAircraft[]
}

export function TrafficGlobe({ aircraft }: TrafficGlobeProps) {
  const containerRef = useRef<HTMLDivElement | null>(null)
  const markerGroupRef = useRef<THREE.Group | null>(null)
  const markerGeometryRef =
    useRef<THREE.SphereGeometry | null>(null)
  const markerMaterialRef =
    useRef<THREE.MeshBasicMaterial | null>(null)

  useEffect(() => {
    if (!containerRef.current) {
      return
    }

    const container = containerRef.current
    const scene = new THREE.Scene()
    const camera = new THREE.PerspectiveCamera(45, 1, 0.1, 1000)
    camera.position.z = 4

    const renderer = new THREE.WebGLRenderer({
      antialias: true,
      alpha: true,
    })

    renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2))
    container.appendChild(renderer.domElement)

    const worldGroup = new THREE.Group()
    scene.add(worldGroup)

    const globeGeometry = new THREE.SphereGeometry(1.4, 64, 64)
    const globeMaterial = new THREE.MeshBasicMaterial({
      color: 0x0f172a,
      wireframe: true,
    })
    const globe = new THREE.Mesh(globeGeometry, globeMaterial)

    const markerGeometry = new THREE.SphereGeometry(0.035, 16, 16)
    const markerMaterial = new THREE.MeshBasicMaterial({
      color: 0x38bdf8,
    })
    const markerGroup = new THREE.Group()

    worldGroup.add(globe, markerGroup)

    markerGroupRef.current = markerGroup
    markerGeometryRef.current = markerGeometry
    markerMaterialRef.current = markerMaterial

    const resize = () => {
      const width = Math.max(container.clientWidth, 1)
      const height = Math.max(container.clientHeight, 1)

      camera.aspect = width / height
      camera.updateProjectionMatrix()
      renderer.setSize(width, height, false)
    }

    resize()

    const resizeObserver = new ResizeObserver(resize)
    resizeObserver.observe(container)

    let animationFrameID = 0

    const animate = () => {
      worldGroup.rotation.y += 0.002
      renderer.render(scene, camera)
      animationFrameID = requestAnimationFrame(animate)
    }

    animate()

    return () => {
      cancelAnimationFrame(animationFrameID)
      resizeObserver.disconnect()

      markerGroup.clear()
      markerGeometry.dispose()
      markerMaterial.dispose()
      globeGeometry.dispose()
      globeMaterial.dispose()

      renderer.dispose()
      renderer.forceContextLoss()

      if (renderer.domElement.parentNode === container) {
        container.removeChild(renderer.domElement)
      }

      markerGroupRef.current = null
      markerGeometryRef.current = null
      markerMaterialRef.current = null
    }
  }, [])

  useEffect(() => {
    const markerGroup = markerGroupRef.current
    const markerGeometry = markerGeometryRef.current
    const markerMaterial = markerMaterialRef.current

    if (!markerGroup || !markerGeometry || !markerMaterial) {
      return
    }

    markerGroup.clear()

    for (const item of aircraft) {
      if (!hasValidCoordinates(item)) {
        continue
      }

      const marker = new THREE.Mesh(
        markerGeometry,
        markerMaterial
      )
      const latitude = THREE.MathUtils.degToRad(item.latitude)
      const longitude = THREE.MathUtils.degToRad(item.longitude)
      const radius = 1.45

      marker.position.set(
        radius * Math.cos(latitude) * Math.sin(longitude),
        radius * Math.sin(latitude),
        radius * Math.cos(latitude) * Math.cos(longitude)
      )

      markerGroup.add(marker)
    }
  }, [aircraft])

  return (
    <div className='h-[600px] w-full overflow-hidden rounded-xl border border-slate-800 bg-black'>
      <div ref={containerRef} className='h-full w-full' />
    </div>
  )
}

function hasValidCoordinates(item: TrafficAircraft): boolean {
  return (
    Number.isFinite(item.latitude) &&
    item.latitude >= -90 &&
    item.latitude <= 90 &&
    Number.isFinite(item.longitude) &&
    item.longitude >= -180 &&
    item.longitude <= 180
  )
}
