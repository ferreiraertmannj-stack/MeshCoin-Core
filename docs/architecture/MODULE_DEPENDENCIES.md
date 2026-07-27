# MESHCOIN CORE

# MODULE DEPENDENCIES

Version 1.0

Architecture Authority Document

---

# PURPOSE

This document defines the official architecture of MeshCoin Core.

Every module has a single responsibility.

Every dependency is intentional.

No AI may change these relationships.

If a change requires violating this document:

STOP.

Explain.

Wait for approval.

---

# SYSTEM OVERVIEW

MeshCoin Core is NOT a monolithic application.

It is composed of multiple isolated systems.

Each system communicates only through defined interfaces.

The architecture is based on:

Low Coupling

High Cohesion

Deterministic State

Distributed Processing

Protocol Isolation

Layer Independence

---

# GLOBAL ARCHITECTURE

                         USER

                           │

                     Flutter Client

                           │

                Application Services

                           │

          ┌─────────────────────────────────┐

          │                                 │

      Wallet Layer                  Mesh Layer

          │                                 │

          └──────────────┬──────────────────┘

                         │

                  Consensus Layer

                         │

                 Blockchain Layer

                         │

                 Storage Layer

                         │

                    Local Database

                         │

                Network Synchronizer

                         │

          Internet / Mesh / PC Nodes

---

# MODULE RESPONSIBILITIES

Every module owns exactly ONE responsibility.

No exceptions.

---

Wallet Layer

Responsible for:

Address generation

Private keys

Public keys

Signing

Balance display

Wallet import/export

Wallet encryption

Wallet recovery

Wallet state

Wallet history

Wallet security

Wallet validation

Wallet NEVER validates blocks.

Wallet NEVER mines.

Wallet NEVER resolves forks.

Wallet NEVER controls networking.

Wallet NEVER modifies blockchain.

---

Blockchain Layer

Responsible for:

Blocks

Transactions

Hash chain

Difficulty

Rewards

Validation

Chain state

Ledger

Genesis

Block verification

Blockchain NEVER controls UI.

Blockchain NEVER controls Mesh.

Blockchain NEVER stores widgets.

Blockchain NEVER knows Flutter.

---

Consensus Layer

Responsible for:

Fork resolution

Conflict detection

Longest chain

Difficulty adjustment

Transaction confirmation

Node agreement

Block acceptance

Chain synchronization

Consensus NEVER accesses UI.

Consensus NEVER draws screens.

Consensus NEVER handles buttons.

Consensus NEVER performs animations.

---

Mesh Layer

Responsible for:

Discovery

Routing

Forwarding

TTL

Packet IDs

Deduplication

Broadcast

Store-and-forward

Neighbor table

Relay

Multi-hop

Offline communication

Mesh NEVER validates blockchain.

Mesh NEVER signs transactions.

Mesh NEVER changes balances.

---

Network Layer

Responsible for:

TCP

UDP

Bluetooth

WiFi Direct

WebSocket

Internet

Connection retries

Timeouts

Compression

Packet transport

Network NEVER validates business rules.

---

Storage Layer

Responsible for:

Persistence

Indexes

Caching

Snapshots

Recovery

Serialization

Storage NEVER validates transactions.

Storage NEVER mines.

Storage NEVER routes packets.

---

Explorer

Responsible only for visualization.

Read-only.

Explorer NEVER changes data.

Explorer NEVER mines.

Explorer NEVER validates.

Explorer NEVER broadcasts.

---

Mining Layer

Responsible for:

Proof of Work

Nonce generation

Difficulty target

Reward proposal

Candidate blocks

Mining NEVER changes wallet balance directly.

Mining NEVER validates transactions.

Mining NEVER resolves consensus.

---

Synchronization Layer

Responsible for:

Node synchronization

Block synchronization

Transaction synchronization

Recovery

Checkpoint

State reconciliation

Synchronization NEVER changes protocol.

Synchronization NEVER invents blocks.

Synchronization NEVER modifies signatures.

---

Nebula Cloud

Responsible for:

Distributed computing

AI tasks

Resource scheduling

Node allocation

Execution control

Nebula NEVER changes blockchain.

Nebula NEVER validates transactions.

Nebula NEVER owns wallet logic.

Nebula is independent.

---

# ALLOWED DEPENDENCIES

Flutter

↓

Application

↓

Wallet

↓

Consensus

↓

Blockchain

↓

Storage

↓

Network

Wallet

↓

Crypto

Consensus

↓

Blockchain

Consensus

↓

Storage

Blockchain

↓

Storage

Blockchain

↓

Crypto

Mesh

↓

Network

Mesh

↓

Storage

Synchronization

↓

Blockchain

Synchronization

↓

Mesh

Synchronization

↓

Network

Nebula

↓

Mesh

Nebula

↓

Scheduler

Nebula

↓

Storage

---

# FORBIDDEN DEPENDENCIES

Wallet

X

Flutter Widgets

Wallet

X

Networking

Wallet

X

Mesh

Wallet

X

Mining

Blockchain

X

Flutter

Blockchain

X

Animations

Blockchain

X

UI

Consensus

X

Flutter

Consensus

X

Storage UI

Consensus

X

Widgets

Mesh

X

Wallet

Mesh

X

Mining

Mesh

X

Consensus

Explorer

X

Blockchain Write

Explorer

X

Mining

Explorer

X

Synchronization

Mining

X

Wallet

Mining

X

UI

Mining

X

Explorer

Nebula

X

Wallet

Nebula

X

Consensus

Nebula

X

Mining

---

# DATA OWNERSHIP

Every data structure has one owner.

Wallet owns wallet state.

Blockchain owns chain state.

Consensus owns network state.

Mesh owns routing tables.

Storage owns persistence.

Network owns sockets.

Nebula owns task scheduling.

Explorer owns visualization.

Ownership MUST NEVER overlap.

---

# SINGLE SOURCE OF TRUTH

Wallet Balance

Owner:

Blockchain

NOT UI

NOT Wallet

NOT Explorer

Current Chain

Owner:

Consensus

NOT Storage

NOT Flutter

Current Connections

Owner:

Mesh

NOT UI

NOT Wallet

Node Identity

Owner:

Wallet

Task Queue

Owner:

Nebula

---

# STATE MACHINE

A transaction follows exactly this path.

Wallet

↓

Sign

↓

Validation

↓

Mempool

↓

Mining

↓

Candidate Block

↓

Consensus

↓

Chain

↓

Synchronization

↓

Wallet Update

Never skip states.

Never reorder states.

---

# FORK STATE

Forks are NORMAL.

Forks are EXPECTED.

Forks are NOT bugs.

Consensus decides.

Wallet waits.

UI reflects.

Never hide forks.

Never delete forks.

Never invent fork resolution.

---

# FAILURE POLICY

Failures must be recoverable.

Never panic.

Never crash intentionally.

Retry.

Recover.

Synchronize.

Continue.

---

# LOGGING POLICY

Every critical event must be logged.

Connection

Synchronization

Mining

Fork

Consensus

Wallet

Recovery

Validation

Storage

Network

Logs are part of architecture.

Never remove logging.

---

# TESTABILITY

Every module must be testable independently.

Wallet tests.

Consensus tests.

Blockchain tests.

Mesh tests.

Storage tests.

Synchronization tests.

Mining tests.

Nebula tests.

---

# FINAL RULE

If an implementation violates module responsibility,

THE IMPLEMENTATION IS WRONG.

NOT THE ARCHITECTURE.

END OF DOCUMENT