# MESHCOIN CORE

# AI_CONTEXT.md

Version 1.0

Persistent Context For Artificial Intelligence

-------------------------------------------------------------------------------

# INTRODUCTION

This document provides permanent architectural context for every AI working on
MeshCoin Core.

It is NOT a coding guide.

It is NOT a prompt.

It is NOT documentation for users.

It exists so every AI understands WHY the project exists before touching any
source code.

Every AI MUST read this document before reading the source code.

-------------------------------------------------------------------------------

# WHAT IS MESHCOIN

MeshCoin is NOT a cryptocurrency.

MeshCoin is NOT "another blockchain".

MeshCoin is NOT a Bitcoin clone.

MeshCoin is NOT an Ethereum clone.

MeshCoin is NOT a wallet application.

MeshCoin is NOT a DeFi platform.

MeshCoin is a decentralized digital infrastructure.

Blockchain is only one subsystem.

The real product is resilient distributed infrastructure.

-------------------------------------------------------------------------------

# PROJECT MISSION

Create a decentralized infrastructure capable of operating with or without
internet.

The network must survive:

Internet outages

Government censorship

ISP failures

Natural disasters

Remote areas

Low-resource hardware

Intermittent connectivity

MeshCoin must continue operating whenever devices can still communicate.

-------------------------------------------------------------------------------

# LONG TERM VISION

MeshCoin will become the communication layer for Nebula Cloud.

Nebula Cloud will provide distributed computing.

MeshCoin will provide:

Identity

Consensus

Synchronization

Distributed trust

Secure communication

Task propagation

Resource discovery

Economic incentives

The blockchain is not the final goal.

The infrastructure is.

-------------------------------------------------------------------------------

# DESIGN PHILOSOPHY

Everything is modular.

Everything is deterministic.

Everything is recoverable.

Everything is distributed.

Nothing depends on centralized servers.

Nothing assumes stable internet.

Nothing assumes high-end hardware.

-------------------------------------------------------------------------------

# HARDWARE TARGET

The project is optimized for:

Android phones

Old smartphones

Samsung devices

Xiaomi devices

ARM processors

Raspberry Pi

Orange Pi

Mini PCs

Desktop PCs

Low-power hardware is a first-class citizen.

Never optimize exclusively for desktop hardware.

-------------------------------------------------------------------------------

# WHY MOBILE FIRST

Millions of smartphones become electronic waste every year.

Those devices already contain:

CPU

GPU

RAM

Flash Storage

WiFi

Bluetooth

Battery

Sensors

MeshCoin transforms those devices into useful network nodes.

-------------------------------------------------------------------------------

# OFFLINE FIRST

The network must always assume that internet may disappear.

Offline operation is NOT an exception.

Offline operation is the default design philosophy.

Internet synchronization is an enhancement.

Never design features that require permanent internet access.

-------------------------------------------------------------------------------

# DISTRIBUTED THINKING

Every node is independent.

Every node may disconnect.

Every node may reconnect later.

Every node must recover.

Every node must synchronize.

No node is permanently trusted.

Trust is earned through protocol validation.

-------------------------------------------------------------------------------

# WHY FORKS EXIST

Forks are not failures.

Forks are expected.

Forks happen naturally in decentralized systems.

Never try to eliminate forks.

Improve fork resolution.

-------------------------------------------------------------------------------

# WHY MESH EXISTS

Mesh networking exists because internet cannot be assumed.

Mesh networking provides:

Discovery

Routing

Forwarding

Delay tolerance

Store-and-forward

Multi-hop communication

Offline communication

Mesh is not an accessory.

Mesh is a fundamental layer.

-------------------------------------------------------------------------------

# WHY BLOCKCHAIN EXISTS

Blockchain exists to provide:

Immutable history

Consensus

Distributed trust

Conflict resolution

Identity

Transaction ordering

It does NOT exist for speculation.

-------------------------------------------------------------------------------

# WHAT AI MUST NEVER ASSUME

Never assume:

Internet exists.

GPS exists.

Cloud exists.

Fast hardware exists.

Permanent storage exists.

Stable power exists.

High bandwidth exists.

Single miner exists.

Single validator exists.

Never assume only one node is mining.

-------------------------------------------------------------------------------

# CURRENT PROJECT STATUS

This project is under active development.

Some systems are intentionally incomplete.

Some systems are experimental.

Some systems are placeholders for future implementations.

Those placeholders are intentional.

Do NOT replace them automatically.

If a subsystem is incomplete:

Ask first.

-------------------------------------------------------------------------------

# KNOWN ENGINEERING CHALLENGES

Current engineering priorities include:

Synchronization stability

Fork resolution

Mesh routing

Connection recovery

Mempool consistency

State reconciliation

Offline synchronization

Node recovery

These are difficult distributed-system problems.

Do not simplify them.

-------------------------------------------------------------------------------

# WHAT LOOKS LIKE A BUG MAY NOT BE A BUG

Distributed systems often behave differently from centralized software.

Examples:

Temporary forks

Duplicate packets

Message retries

Delayed synchronization

Out-of-order arrival

Eventually consistent state

These behaviors may be expected.

Never "fix" them without understanding the protocol.

-------------------------------------------------------------------------------

# CODE MODIFICATION PHILOSOPHY

Before changing any code ask:

Why does this code exist?

What protocol depends on it?

What state machine depends on it?

Who calls it?

Who receives its output?

What invariant does it preserve?

If these questions cannot be answered,

do not modify the code.

-------------------------------------------------------------------------------

# ARCHITECTURAL PRIORITIES

Priority 1

Correctness

Priority 2

Determinism

Priority 3

Recoverability

Priority 4

Compatibility

Priority 5

Performance

Never invert this order.

-------------------------------------------------------------------------------

# AI BEHAVIOR

The AI should behave like:

A senior distributed systems engineer.

Not like:

A code generator.

Not like:

An automatic refactoring tool.

Not like:

A linter.

Not like:

A formatter.

-------------------------------------------------------------------------------

# IF YOU DISAGREE

If you believe architecture should change:

Do not implement.

Explain.

Present evidence.

Wait.

-------------------------------------------------------------------------------

# FINAL MESSAGE

You are helping to build infrastructure.

Infrastructure is preserved through discipline.

Not through shortcuts.

Architecture comes first.

Code comes second.

Always preserve the architecture.

END OF DOCUMENT