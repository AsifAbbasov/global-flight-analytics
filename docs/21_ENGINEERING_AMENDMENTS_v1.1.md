# DOCUMENT 21

# ENGINEERING AMENDMENTS

## Global Flight Analytics

**Version:** 1.1

**Status:** Approved

---

# 1. Purpose

This document records engineering improvements to the project architecture adopted after analyzing existing open aviation platforms, three-dimensional map visualization solutions, air-traffic rendering engines, and modern WebGL application-development practices.

This document does not replace the previously approved architecture.

It extends the existing architecture with engineering decisions intended to:

- improve performance;
- improve scalability;
- improve user interaction;
- improve visualization quality;
- reduce memory consumption;
- prepare the platform for future development.

---

# 2. Scope

The changes apply exclusively to the following subsystems:

- Frontend Rendering Engine
- Three.js Engine
- Aircraft Rendering
- User Interaction
- Navigation
- Scene Management
- Performance Optimization

This document does not change:

- Product Vision;
- System Architecture;
- Domain Model;
- Backend Architecture;
- Database Design;
- API Contract;
- Data Pipeline;
- Security Architecture.

All existing documents remain valid.

---

# 3. Design Principles

## 3.1 Independent Implementation

Engineering ideas were adopted after analyzing third-party solutions.

No third-party source code is used.

All implementations are created exclusively as original project code.

---

## 3.2 Backend First

The Backend remains the only source of truth.

The Frontend never accesses external aviation services directly.

All external data sources pass through the Backend.

---

## 3.3 Scalable Rendering

Every visualization component must be able to work with tens of thousands of simultaneously displayed objects.

All algorithms must account for scaling.

---

## 3.4 Explicit Resource Management

Every created WebGL resource must have a controlled lifecycle.

Geometry, Material, Texture, and Buffer objects must not remain in memory after their owning object is removed.

---

## 3.5 Progressive Enhancement

Every additional capability must improve the user experience without breaking the basic system workflow.

---

# 4. Rendering Engine Improvements

## 4.1 Invisible Hitboxes

Every aircraft receives a separate invisible selection object.

The Hitbox is used exclusively for user interaction.

The selection area may differ from the visible aircraft-model size.

### Benefits

- reliable object selection;
- improved interaction usability;
- fewer incorrect clicks.

Status

Approved

---

## 4.2 Raycaster Selection

Object selection is performed exclusively through Three.js Raycaster.

After an object is selected, the system must:

- identify the aircraft;
- identify the airport;
- open the information card;
- activate visual highlighting;
- prepare Follow Mode.

Status

Approved

---

## 4.3 Selection Highlight

The selected aircraft must be visually distinguishable.

Allowed techniques:

- color changes;
- an outer outline;
- additional glow;
- opacity changes;
- scaling.

Status

Approved

---

## 4.4 Sprite Glow

Sprite Glow is used instead of additional Mesh objects.

Reasons:

- fewer draw calls;
- less Geometry;
- lower memory usage;
- higher performance.

Status

Approved

---

## 4.5 Resource Disposal Manager

A centralized resource-disposal manager is created.

It is responsible for releasing:

- Geometry;
- Material;
- Texture;
- Sprite;
- BufferGeometry;
- RenderTarget;
- Labels.

No object may be removed without releasing its resources.

Status

Approved

---

## 4.6 Batched Updates

LOD processing and scene updates are performed in batches.

The system must not recalculate every object on every frame.

Status

Approved

---

# 5. Aircraft Visualization

## 5.1 Trail Buffer

Each aircraft's movement history is stored in a ring buffer.

The buffer reuses allocated memory.

Creating new arrays on every update is prohibited.

Status

Approved

---

## 5.2 Trail Level of Detail

The number of route points depends on distance from the camera.

Nearby aircraft display the full route.

Distant aircraft use a reduced trajectory.

Status

Approved

---

## 5.3 Aircraft Orientation

When data is available, the model must display:

- Heading;
- Pitch;
- Bank.

When some parameters are unavailable, only available values are used.

Status

Approved

---

## 5.4 Follow Aircraft

After selecting an aircraft, the user may enable automatic tracking.

The camera remains attached to the selected object.

Status

Approved

---

## 5.5 Stale Aircraft Cleanup

Aircraft that have not received an update within the configured period are removed automatically.

Removal includes:

- model;
- trail;
- label;
- glow;
- internal data structures.

Status

Approved

---

# 6. Navigation

## 6.1 URL State

Application state must be restorable from the URL.

Supported state:

- aircraft;
- airport;
- region.

Status

Approved

---

## 6.2 Deep Linking

Every system object has a permanent link.

Examples:

- aircraft;
- airport;
- region.

Opening the link must fully restore interface state.

Status

Approved

---

# 7. Airport Improvements

## 7.1 Airport Search Ranking

Airport search uses a ranking system.

Priority is determined by:

- exact ICAO match;
- exact IATA match;
- airport name;
- airport popularity.

Status

Approved

---

## 7.2 Aircraft Photo Service

The aircraft card may display an aircraft photograph.

The photo source is a separate integration service.

The system must continue to operate unchanged when no photo is available.

Status

Approved

---

# 8. Performance Improvements

## 8.1 Focused Radius Loading

Data is loaded only around the current observation region.

Loading the entire global airspace is prohibited.

Status

Approved

---

## 8.2 Adaptive Update Strategy

Update frequency depends on:

- device performance;
- object count;
- distance from the camera.

Status

Approved

---

## 8.3 Terrain Streaming

Terrain tiles are loaded only when required.

Distant areas are released from memory automatically.

Status

Future

---

## 8.4 Terrain Level of Detail

Terrain detail depends on distance from the camera.

Status

Future

---

# 9. User Experience

## 9.1 Aircraft Information Card

After selecting an aircraft, the system displays one unified information card.

The card aggregates information from several domain entities.

In the future, it may include:

- Aircraft;
- Flight;
- Route;
- Aircraft Profile;
- photographs;
- model specifications;
- flight history.

Status

Approved

---

## 9.2 Smooth Interaction

All user actions must operate without a complete scene redraw.

Status

Approved

---

# 10. Version 2 Features

The following capabilities are outside the minimum viable product.

## Recording

Record scene state.

Status

Future

---

## Replay

Replay previously recorded state.

Status

Future

---

## Surf Mode

Move freely through airspace.

Status

Future

---

## Timeline Playback

Move through a timeline.

Status

Future

---

# 11. Explicitly Rejected Decisions

The following decisions are intentionally rejected:

- direct Frontend access to third-party aviation services;
- storing all history only in the browser;
- operating without a project-owned Backend;
- operating without a project-owned domain layer;
- operating without a project-owned data model;
- operating without centralized WebGL resource management.

---

# 12. Compatibility

This document is fully compatible with the following project documents:

- DOCUMENT 01 — Product Vision
- DOCUMENT 02 — System Architecture
- DOCUMENT 03 — Domain Model
- all subsequent architecture documents

This document extends the architecture and does not change previously approved decisions.

---

# 13. Engineering Decisions Adopted

After engineering analysis of existing open aviation platforms, the following decisions were adopted for the Global Flight Analytics architecture:

- Invisible Hitboxes;
- Raycaster Selection;
- Selection Highlight;
- Sprite Glow;
- Resource Disposal Manager;
- Trail Buffer;
- Trail Level of Detail;
- Aircraft Orientation;
- Follow Aircraft;
- Stale Aircraft Cleanup;
- URL State;
- Deep Linking;
- Airport Search Ranking;
- Aircraft Photo Integration;
- Focused Radius Loading;
- Adaptive Update Strategy;
- Batched Scene Updates;
- Terrain Streaming;
- Terrain Level of Detail;
- Recording and Replay as Version 2 functionality.

All listed decisions are implemented exclusively by the project development team, integrated into the existing architecture, and do not involve copying third-party implementations.

---

# 14. Amendment Summary

This document closes the architecture phase of Global Flight Analytics.

After approval, the system architecture is considered sufficient for transition to implementation.

Further architecture changes are allowed only through new amendment documents or Architecture Decision Records.

All subsequent work must follow the previously approved project documentation and this document.

---

# 15. Engineering Decision Matrix

This section defines the final status of every engineering initiative adopted after analysis of existing open aviation platforms.

| Engineering Decision       | Status   | Priority | Planned Release |
| -------------------------- | -------- | -------- | --------------- |
| Invisible Hitboxes         | Approved | High     | MVP             |
| Raycaster Selection        | Approved | High     | MVP             |
| Selection Highlight        | Approved | High     | MVP             |
| Sprite Glow                | Approved | Medium   | MVP             |
| Trail Buffer               | Approved | High     | MVP             |
| Trail Level of Detail      | Approved | High     | MVP             |
| Aircraft Orientation       | Approved | Medium   | MVP             |
| Follow Aircraft            | Approved | Medium   | MVP             |
| Stale Aircraft Cleanup     | Approved | High     | MVP             |
| Resource Disposal Manager  | Approved | High     | MVP             |
| URL State Restoration      | Approved | High     | MVP             |
| Deep Linking               | Approved | Medium   | MVP             |
| Airport Search Ranking     | Approved | Medium   | MVP             |
| Aircraft Photo Integration | Approved | Low      | Version 1.1     |
| Focused Radius Loading     | Approved | High     | MVP             |
| Adaptive Update Strategy   | Approved | High     | MVP             |
| Batched Scene Updates      | Approved | High     | MVP             |
| Terrain Streaming          | Future   | Low      | Version 2       |
| Terrain Level of Detail    | Future   | Low      | Version 2       |
| Recording                  | Future   | Low      | Version 2       |
| Replay                     | Future   | Low      | Version 2       |
| Surf Mode                  | Future   | Low      | Version 2       |

---

# 16. Expected Engineering Impact

## 16.1 Rendering Performance

The approved changes are expected to provide:

- fewer Draw Calls;
- lower graphics-memory usage;
- fewer JavaScript memory allocations;
- fewer Geometry recreations;
- fewer Material recreations;
- lower Garbage Collector load;
- more stable frame times;
- support for significantly more simultaneously active aircraft.

---

## 16.2 Scalability

The architecture must support future development without changing the project's fundamental structure.

Expected scaling directions:

- increase the number of simultaneously displayed aircraft;
- increase the number of airports;
- expand supported regions;
- add new data sources;
- introduce new analytical modules;
- introduce historical analysis;
- introduce air-traffic forecasting.

---

## 16.3 User Experience

After implementation, users should receive:

- more accurate aircraft selection;
- reliable scene interaction;
- more informative object cards;
- convenient navigation through permanent links;
- improved tracking of a selected aircraft;
- smoother interface operation.

---

## 16.4 Maintainability

The new engineering decisions are intended to improve project maintainability.

The architecture must provide:

- low component coupling;
- module reuse;
- independent testing;
- gradual capability expansion.

---

# 17. Engineering Guidelines

This section defines mandatory implementation requirements.

## 17.1 Resource Lifecycle

Every created visualization object must have a fully controlled lifecycle.

Creating an object without subsequently releasing its resources is prohibited.

---

## 17.2 Rendering Rules

Every new visual capability must be evaluated for:

- performance;
- memory consumption;
- scalability;
- Draw Call count;
- object count.

---

## 17.3 Data Flow Rules

All user actions operate only on the system's internal data model.

The Frontend does not communicate directly with external services.

All data passes through the Backend.

---

## 17.4 Backward Compatibility

New engineering decisions must not break compatibility with the existing architecture.

Changes must integrate without requiring previously approved documents to be rewritten.

---

# 18. Final Engineering Statement

This document closes the architecture phase of Global Flight Analytics.

After approval, the architecture is considered complete and ready for transition to implementation.

All further architecture changes are allowed exclusively through:

- Engineering Amendments;
- Architecture Decision Records;
- Version Documents.

Software development must follow the approved project architecture documentation.

This document supplements the existing architecture and does not replace previously approved documents.

# 19. Architectural Constraints

The following constraints must be observed during implementation.

## Backend

- The Backend is the only source of truth.
- The Frontend does not access third-party services directly.
- All external APIs pass through the Backend.

## Frontend

- Business logic must not be placed inside user-interface components.
- The Three.js Engine must be isolated from React components.
- User-interface components must not manage the scene lifecycle.

## Rendering

- New Geometry objects must not be created on every frame.
- New Material objects must not be created on every frame.
- Texture objects must not be created without control.
- Every created resource must be released.

## Performance

- Every new capability must be evaluated for performance.
- Scalability takes priority over visual effects.

# 20. Out of Scope

The following capabilities are outside the minimum viable product:

- user system;
- authorization;
- favorite aircraft;
- favorite airports;
- cloud-based user settings;
- collaboration;
- route sharing;
- notifications;
- air-traffic forecasting;
- machine learning;
- a project-owned ADS-B receiver network;
- commercial aviation data.

# 21. Amendment History

| Version | Date                   | Description                                                    |
| ------- | ---------------------- | -------------------------------------------------------------- |
| 1.0     | Initial Architecture   | Initial documentation                                          |
| 1.1     | Engineering Amendments | Rendering, LOD, Selection, URL State, Performance Improvements |
