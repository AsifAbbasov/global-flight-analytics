'use client'

import { useEffect, useRef } from 'react'
import * as THREE from 'three'

import type { TrafficAircraft } from '@/types/traffic'

interface TrafficGlobeProps {
  aircraft: TrafficAircraft[]
}

export function TrafficGlobe({ aircraft }: TrafficGlobeProps) {
  const containerRef = useRef<HTMLDivElement | null>(null)

  useEffect(() => {
    if (!containerRef.current) {
      return
    }

    const container = containerRef.current
    const width = container.clientWidth
    const height = container.clientHeight

    const scene = new THREE.Scene()

    const camera = new THREE.PerspectiveCamera(45, width / height, 0.1, 1000)
    camera.position.z = 4

    const renderer = new THREE.WebGLRenderer({
      antialias: true,
      alpha: true,
    })

    renderer.setSize(width, height)
    renderer.setPixelRatio(window.devicePixelRatio)
    container.appendChild(renderer.domElement)

    const globeGeometry = new THREE.SphereGeometry(1.4, 64, 64)
    const globeMaterial = new THREE.MeshBasicMaterial({
      color: 0x0f172a,
      wireframe: true,
    })

    const globe = new THREE.Mesh(globeGeometry, globeMaterial)
    scene.add(globe)

    const markerMaterial = new THREE.MeshBasicMaterial({
      color: 0x38bdf8,
    })

    aircraft.forEach(item => {
      const markerGeometry = new THREE.SphereGeometry(0.035, 16, 16)
      const marker = new THREE.Mesh(markerGeometry, markerMaterial)

      const latitude = THREE.MathUtils.degToRad(item.latitude)
      const longitude = THREE.MathUtils.degToRad(item.longitude)

      const radius = 1.45

      marker.position.set(
        radius * Math.cos(latitude) * Math.sin(longitude),
        radius * Math.sin(latitude),
        radius * Math.cos(latitude) * Math.cos(longitude)
      )

      scene.add(marker)
    })

    let animationFrameId = 0

    function animate() {
      globe.rotation.y += 0.002
      renderer.render(scene, camera)
      animationFrameId = requestAnimationFrame(animate)
    }

    animate()

    return () => {
      cancelAnimationFrame(animationFrameId)
      renderer.dispose()
      container.removeChild(renderer.domElement)
    }
  }, [aircraft])

  return (
    <div className='h-[600px] w-full overflow-hidden rounded-xl border border-slate-800 bg-black'>
      <div ref={containerRef} className='h-full w-full' />
    </div>
  )
}
