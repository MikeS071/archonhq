# PAPERCLIP_CONNECTOR_SPEC.md

## Purpose
Project JouleWork task/approval/fleet state into Paperclip-compatible operator workflow surfaces.

## Use Paperclip for
- governance UX
- approval projection
- org/team representation
- ticket-like work items
- budget/trace views
- heartbeat/status projections

## Do not use Paperclip for
- durable task truth
- lease truth
- event truth
- ledger truth
- reliability truth

## Connector responsibilities
- sync workspace summaries
- sync approval queue projections
- sync ticket/task summaries
- sync fleet and heartbeat summaries
- sync settlement/reliability widgets or metrics payloads
