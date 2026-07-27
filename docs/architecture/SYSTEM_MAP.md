# MESHCOIN CORE

# SYSTEM MAP

Version 1.0

Official Project Navigation Guide

-------------------------------------------------------------------------------

# PURPOSE

This document is the official navigation map of MeshCoin Core.

It exists to explain the purpose of every major module before an AI attempts to
read or modify the source code.

Source code alone is NOT sufficient to understand the architecture.

Always read this document first.

-------------------------------------------------------------------------------

# PROJECT STRUCTURE

MeshCoin is divided into independent domains.

Each domain owns one responsibility.

No domain may absorb responsibilities from another.

-------------------------------------------------------------------------------

ROOT

The project root contains only:

Project configuration

Documentation

Build configuration

Development tools

Licenses

README

Roadmap

Architecture

No business logic should exist here.

-------------------------------------------------------------------------------

/docs

Purpose

Official documentation.

Contains:

Architecture

Protocols

Roadmaps

Design decisions

Engineering rules

AI manuals

Never place executable business logic inside /docs.

-------------------------------------------------------------------------------

/flutter

Purpose

Entire mobile application.

Contains:

User Interface

Widgets

Navigation

Pages

Animations

State presentation

Flutter must NEVER own blockchain logic.

Flutter must NEVER own consensus.

Flutter must NEVER own mining.

Flutter must NEVER own networking protocols.

Flutter is presentation only.

-------------------------------------------------------------------------------

/wallet

Purpose

Wallet subsystem.

Responsibilities

Generate wallets

Store wallets

Encrypt wallets

Unlock wallets

Sign transactions

Export wallets

Import wallets

Recover wallets

Wallet NEVER validates blockchain.

Wallet NEVER resolves forks.

Wallet NEVER controls synchronization.

Wallet NEVER controls networking.

-------------------------------------------------------------------------------

/blockchain

Purpose

Blockchain engine.

Responsibilities

Blocks

Transactions

Ledger

Validation

Difficulty

Rewards

Hash chain

Genesis

Block verification

Blockchain NEVER knows Flutter.

Blockchain NEVER knows UI.

Blockchain NEVER knows animations.

-------------------------------------------------------------------------------

/consensus

Purpose

Network agreement.

Responsibilities

Fork resolution

Chain selection

Confirmation

Validation

Difficulty adjustment

Conflict resolution

Consensus NEVER controls Wallet.

Consensus NEVER controls UI.

Consensus NEVER controls Mesh.

-------------------------------------------------------------------------------

/mesh

Purpose

Peer-to-peer communication.

Responsibilities

Discovery

Routing

Packet forwarding

Neighbor tables

TTL

Store-and-forward

Broadcast

Relay

Offline communication

Mesh NEVER changes wallet balances.

Mesh NEVER validates transactions.

Mesh NEVER mines.

-------------------------------------------------------------------------------

/network

Purpose

Transport layer.

Responsibilities

TCP

UDP

Bluetooth

WiFi Direct

WebSockets

Internet

Compression

Timeouts

Retries

Network NEVER contains blockchain rules.

-------------------------------------------------------------------------------

/crypto

Purpose

Cryptographic primitives.

Responsibilities

Hashes

Keys

Signatures

Encryption

Verification

Random generation

Crypto NEVER performs networking.

Crypto NEVER performs mining.

Crypto NEVER owns blockchain state.

-------------------------------------------------------------------------------

/mining

Purpose

Proof of Work.

Responsibilities

Nonce generation

Difficulty verification

Candidate blocks

Reward proposal

Mining NEVER updates balances.

Mining NEVER validates transactions.

Mining NEVER resolves consensus.

-------------------------------------------------------------------------------

/storage

Purpose

Persistent storage.

Responsibilities

Database

Serialization

Snapshots

Indexes

Recovery

Caching

Storage NEVER validates blockchain.

Storage NEVER mines.

Storage NEVER routes packets.

-------------------------------------------------------------------------------

/sync

Purpose

Synchronization engine.

Responsibilities

Block sync

Transaction sync

Recovery

State reconciliation

Checkpoint

Resync

Synchronization NEVER creates transactions.

Synchronization NEVER mines.

Synchronization NEVER changes consensus rules.

-------------------------------------------------------------------------------

/explorer

Purpose

Visualization.

Responsibilities

Read blockchain

Display blocks

Display transactions

Display addresses

Statistics

Explorer NEVER modifies blockchain.

Explorer NEVER validates blocks.

Explorer NEVER mines.

Explorer NEVER broadcasts packets.

-------------------------------------------------------------------------------

/nebula

Purpose

Distributed computing.

Responsibilities

Task scheduling

Resource allocation

AI execution

Distributed jobs

Node orchestration

Nebula NEVER validates blockchain.

Nebula NEVER owns wallet logic.

Nebula NEVER mines.

-------------------------------------------------------------------------------

# DEPENDENCY DIRECTION

Allowed flow

UI

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

Forbidden flow

Blockchain

↓

Flutter

Consensus

↓

Widgets

Mining

↓

Wallet

Mesh

↓

Consensus

Explorer

↓

Blockchain Write

-------------------------------------------------------------------------------

# MODULE OWNERSHIP

Wallet owns

Wallet State

------------------------------------------------

Blockchain owns

Ledger

------------------------------------------------

Consensus owns

Current Chain

------------------------------------------------

Mesh owns

Routing Table

------------------------------------------------

Storage owns

Persistence

------------------------------------------------

Network owns

Sockets

------------------------------------------------

Nebula owns

Distributed Tasks

------------------------------------------------

Explorer owns

Visualization

-------------------------------------------------------------------------------

# FILE MODIFICATION POLICY

Before modifying a file the AI must answer

Why does this file exist?

Who owns this module?

Which protocol depends on it?

Who calls this code?

Who receives its output?

What invariants does it preserve?

If these questions cannot be answered

STOP.

-------------------------------------------------------------------------------

# HIGH RISK FILES

The following files are considered critical.

Blockchain core

Consensus engine

Mining engine

Wallet generation

Synchronization

Mesh routing

Packet serialization

Cryptographic functions

Node state machine

These files require maximum caution.

-------------------------------------------------------------------------------

# SAFE FILES

Documentation

Comments

README

Roadmaps

Localization

Translations

Minor UI fixes

Non-business styling

-------------------------------------------------------------------------------

# BEFORE ADDING NEW FILES

Ask:

Is a similar module already present?

Does this responsibility already belong to another module?

Can the feature be isolated?

Does this create duplicated logic?

If yes

Do not create the file.

-------------------------------------------------------------------------------

# BEFORE MOVING CODE

Moving code is architecture.

Architecture belongs only to the architect.

Never move code automatically.

-------------------------------------------------------------------------------

# BEFORE DELETING CODE

Never delete code because it "looks unused".

Distributed systems often contain:

Recovery paths

Fallback logic

Compatibility code

Migration code

Experimental features

Ask first.

-------------------------------------------------------------------------------

# AI NAVIGATION RULE

Never navigate the project by guessing.

Always use:

Architecture

System Map

Protocols

Only then

Read code.

-------------------------------------------------------------------------------

# FINAL MESSAGE

Understanding the architecture is mandatory.

Reading the code is not enough.

Architecture defines meaning.

Code only implements it.

END OF DOCUMENT