# Akilix LLM Architecture

## Purpose

LLM integration assists analysis.

The model is not an evidence source.

---

# Modes

Akilix defines four conceptual modes.

## OFF

No LLM use.

## LOCAL

Inference occurs locally.

This is the natural default for sensitive investigations.

## REMOTE-REDACTED

Selected material is transformed/redacted according to explicit policy before remote transmission.

## REMOTE-EXPLICIT

Raw selected material may be transmitted only after explicit operator approval for the operation.

---

# Evidence broker

Models should not receive arbitrary unrestricted workbook filesystem access.

Prefer an evidence broker exposing controlled operations such as:

    get_artifact(...)
    get_evidence_metadata(...)
    get_invocation(...)
    search_timeline(...)
    search_records(...)

This makes context selection inspectable and supports source traceability.

---

# Epistemic classes

LLM-assisted claims must distinguish:

## FACT

A direct statement grounded in cited evidence.

## INFERENCE

A reasoned conclusion based on one or more facts.

## HYPOTHESIS

A proposition requiring additional corroboration.

These classifications should be machine-readable.

---

# Source references

Factual statements should carry source references where possible.

Examples:

- evidence ID
- artifact ID
- record ID
- file path
- byte offset
- timeline event
- invocation ID

---

# Model provenance

LLM-assisted output should be capable of recording:

- model identifier
- model version/digest where available
- local/remote mode
- generation timestamp
- source context references
- redaction policy where applicable

---

# Remote operation

Remote providers are optional adapters.

Core Akilix operation must not depend on a remote model.

`REMOTE-EXPLICIT` must not silently become persistent consent.

Each selected transmission should remain inspectable.

---

# Offline operation

Local LLM support should be designed to operate in disconnected laboratories with locally stored models.

Model download is an administrative activity, not an implicit workbook action.
