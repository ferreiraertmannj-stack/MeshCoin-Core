# MESHCOIN CORE

# CODEBASE INDEX

Version 1.0

Official Source Code Navigation Index

-------------------------------------------------------------------------------

# PURPOSE

This document is the official index of the MeshCoin Core codebase.

Its purpose is to describe every important file, module and package before an AI
attempts to modify the source code.

Source code is implementation.

This document is meaning.

Every AI must consult this index before modifying any file.

-------------------------------------------------------------------------------

# FILE STATUS LEGEND

★★★★★ Core Critical

★★★★ High Critical

★★★ Medium

★★ Low

★ Utility

-------------------------------------------------------------------------------

# FILE MATURITY

Production

Stable

Experimental

Prototype

Deprecated

Future

-------------------------------------------------------------------------------

# MODIFICATION LEVEL

LOCKED

Only the architect may approve changes.

RESTRICTED

Changes require justification.

NORMAL

May be modified.

FREE

Safe for improvements.

-------------------------------------------------------------------------------

# FILE TEMPLATE

Every important source file must be documented using the following format.

-------------------------------------------------------------------------------

FILE

lib/path/example.dart

Purpose

Short description.

Responsibilities

•

•

•

Called By

•

•

Calls

•

•

Owns

•

Never Owns

•

•

State

Production

Criticality

★★★★★

Modification

Restricted

Protocols

CORE_PROTOCOL.md

Invariants

BC-001

BC-005

BC-010

Dependencies

crypto.dart

storage.dart

Never Depend On

Flutter

Widgets

UI

Known Risks

Describe risks.

Known Technical Debt

Describe debt.

Future Evolution

Describe expected evolution.

-------------------------------------------------------------------------------

# DIRECTORY INDEX

The following sections should be completed as the project evolves.

-------------------------------------------------------------------------------

## lib/

Purpose

Application source code.

Contains

Business logic.

Flutter UI.

Services.

Networking.

Protocols.

Blockchain.

-------------------------------------------------------------------------------

## lib/core/

Purpose

Core business rules.

Criticality

★★★★★

Modification

Restricted

Rules

Never place UI inside core.

Never place widgets inside core.

-------------------------------------------------------------------------------

## lib/core/blockchain/

Purpose

Blockchain engine.

Owns

Blocks

Transactions

Validation

Ledger

Difficulty

Never Owns

UI

Networking

Consensus

Expected Files

block.dart

blockchain.dart

transaction.dart

ledger.dart

validator.dart

-------------------------------------------------------------------------------

## lib/core/consensus/

Purpose

Consensus engine.

Criticality

★★★★★

Rules

Only consensus decides chain validity.

Never duplicate consensus logic elsewhere.

-------------------------------------------------------------------------------

## lib/core/mining/

Purpose

Mining subsystem.

Criticality

★★★★★

Responsibilities

Proof of Work

Nonce

Difficulty

Candidate blocks

Never

Update balances directly.

-------------------------------------------------------------------------------

## lib/core/wallet/

Purpose

Wallet subsystem.

Criticality

★★★★★

Responsibilities

Identity

Keys

Signatures

Encryption

Never

Own blockchain state.

-------------------------------------------------------------------------------

## lib/core/mesh/

Purpose

Mesh networking.

Criticality

★★★★★

Responsibilities

Routing

Relay

Broadcast

TTL

Neighbors

Offline packets

Never

Modify blockchain.

-------------------------------------------------------------------------------

## lib/core/network/

Purpose

Transport layer.

Criticality

★★★★

Responsibilities

TCP

UDP

Bluetooth

WebSocket

Internet

Retries

Timeouts

-------------------------------------------------------------------------------

## lib/core/storage/

Purpose

Persistent storage.

Criticality

★★★★

Responsibilities

Database

Serialization

Snapshots

Recovery

-------------------------------------------------------------------------------

## lib/core/sync/

Purpose

Synchronization engine.

Criticality

★★★★★

Responsibilities

Import blocks

Import transactions

Recovery

Consensus synchronization

-------------------------------------------------------------------------------

## lib/core/crypto/

Purpose

Cryptographic primitives.

Criticality

★★★★★

Responsibilities

Hashes

Keys

Encryption

Signatures

Verification

-------------------------------------------------------------------------------

## lib/core/explorer/

Purpose

Read-only blockchain visualization.

Criticality

★★★

Never

Modify blockchain.

-------------------------------------------------------------------------------

## lib/core/nebula/

Purpose

Nebula Cloud integration.

Criticality

★★★★★

Responsibilities

Distributed execution

Task scheduling

Resource allocation

Edge computing

-------------------------------------------------------------------------------

## lib/services/

Purpose

Application services.

Criticality

★★★

Services connect UI and Core.

Never implement blockchain rules here.

-------------------------------------------------------------------------------

## lib/widgets/

Purpose

Reusable Flutter widgets.

Criticality

★

Never implement business rules.

-------------------------------------------------------------------------------

## lib/screens/

Purpose

Application pages.

Criticality

★

Presentation only.

-------------------------------------------------------------------------------

## lib/models/

Purpose

DTOs and immutable data models.

Criticality

★★★

Rules

Models must not contain business logic.

-------------------------------------------------------------------------------

## lib/utils/

Purpose

Pure helper utilities.

Criticality

★★

Utilities should be deterministic.

-------------------------------------------------------------------------------

# SOURCE FILE CLASSIFICATION

Every important file must belong to exactly one category.

Blockchain

Consensus

Wallet

Mining

Mesh

Synchronization

Networking

Storage

Crypto

Explorer

Nebula

Flutter

Widgets

Utilities

Testing

Documentation

-------------------------------------------------------------------------------

# CODE OWNERSHIP

Every file has exactly one owner.

Never create multiple owners for the same responsibility.

-------------------------------------------------------------------------------

# FILE DEPENDENCY RULE

Dependencies always point downward.

Higher layers depend on lower layers.

Lower layers never depend on higher layers.

-------------------------------------------------------------------------------

# BEFORE MODIFYING A FILE

Answer:

Why does this file exist?

Who owns it?

Who calls it?

Who depends on it?

Which protocol does it implement?

Which invariant does it protect?

What breaks if it changes?

If unknown,

STOP.

-------------------------------------------------------------------------------

# BEFORE CREATING A FILE

Answer:

Does a similar file already exist?

Does another module already own this responsibility?

Can the current architecture support this feature without creating a new file?

-------------------------------------------------------------------------------

# BEFORE DELETING A FILE

Determine:

Is it referenced?

Is it part of recovery?

Is it compatibility code?

Is it future architecture?

Never delete because it appears unused.

-------------------------------------------------------------------------------

# FILE REVIEW CHECKLIST

Every modified file must answer:

Purpose preserved?

Architecture preserved?

Protocol preserved?

API preserved?

Invariants preserved?

Tests still valid?

Logging preserved?

Recovery preserved?

-------------------------------------------------------------------------------

# ENGINEERING NOTE

A file is more than source code.

It is part of the architecture.

Changing one file may affect dozens of protocols.

Always understand before modifying.

-------------------------------------------------------------------------------

# FUTURE IMPROVEMENT

As the project grows, every critical file should receive its own section in this
document, documenting:

- Exact path
- Owner
- Responsibility
- Public API
- Internal dependencies
- External dependencies
- State machine
- Protocols
- Invariants
- Unit tests
- Integration tests
- Known bugs
- Performance considerations
- Security considerations
- Future roadmap

This document is expected to evolve continuously.

-------------------------------------------------------------------------------

# FINAL MESSAGE

Understanding the codebase is mandatory.

Modifying code without understanding ownership is forbidden.

Architecture always has priority over implementation.

END OF DOCUMENT