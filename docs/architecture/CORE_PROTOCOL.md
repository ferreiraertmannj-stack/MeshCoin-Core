# MESHCOIN CORE

# CORE PROTOCOL

Version 1.0

Official Network Protocol Specification

---

# PURPOSE

This document defines every official protocol used by MeshCoin Core.

Every implementation MUST follow this specification.

No implementation may invent new states.

No implementation may skip states.

No implementation may reorder states.

This protocol is deterministic.

---

# NETWORK PHILOSOPHY

MeshCoin is a Distributed State Machine.

Everything in the network is represented by states.

Nodes never "guess".

Nodes never "assume".

Nodes always validate.

Nodes always synchronize.

Nodes always converge.

---

# NETWORK COMPONENTS

The network is composed of:

Wallet Nodes

Mining Nodes

Relay Nodes

PC Full Nodes

Explorer Nodes

Nebula Nodes

Every node has one responsibility.

---

# NODE LIFECYCLE

Node Created

↓

Wallet Loaded

↓

Identity Verified

↓

Storage Loaded

↓

Blockchain Loaded

↓

Network Started

↓

Mesh Started

↓

Synchronization Started

↓

Consensus Started

↓

Ready

Nodes NEVER skip initialization.

---

# NODE STATES

BOOTING

INITIALIZING

LOADING_STORAGE

LOADING_CHAIN

CONNECTING

SYNCING

READY

MINING

RELAYING

OFFLINE

RECOVERING

ERROR

Every node must always know its current state.

---

# NODE TRANSITIONS

BOOTING

↓

INITIALIZING

↓

LOADING_STORAGE

↓

LOADING_CHAIN

↓

CONNECTING

↓

SYNCING

↓

READY

↓

MINING

If synchronization fails

↓

RECOVERING

↓

SYNCING

↓

READY

Never jump directly to READY.

---

# WALLET STATE MACHINE

Wallet Created

↓

Key Generated

↓

Encrypted

↓

Stored

↓

Loaded

↓

Unlocked

↓

Ready

Wallet NEVER creates transactions before Ready.

---

# TRANSACTION LIFECYCLE

Transaction Created

↓

Local Validation

↓

Signed

↓

Hash Generated

↓

Added To Local Mempool

↓

Broadcast

↓

Neighbor Validation

↓

Global Mempool

↓

Mining Candidate

↓

Block Proposal

↓

Consensus

↓

Block Accepted

↓

Synchronization

↓

Wallet Balance Updated

↓

History Updated

Never skip validation.

Never update balance before consensus.

---

# MEMPOOL

Every transaction has one state.

NEW

↓

LOCAL_VALIDATED

↓

SIGNED

↓

BROADCASTED

↓

RECEIVED

↓

GLOBAL_VALIDATED

↓

WAITING_MINING

↓

MINED

↓

CONFIRMED

↓

FINALIZED

Rejected transactions are removed only with reason.

---

# BLOCK LIFECYCLE

Candidate

↓

Hashing

↓

Difficulty Check

↓

Local Validation

↓

Broadcast

↓

Peer Validation

↓

Consensus

↓

Accepted

↓

Persisted

↓

Synchronized

Never persist before validation.

---

# BLOCK VALIDATION

Every received block must validate:

Previous Hash

Timestamp

Difficulty

Nonce

Reward

Merkle Root

Transactions

Signatures

Chain Height

If one validation fails

Reject.

Never partially accept.

---

# CHAIN SYNCHRONIZATION

Synchronization is deterministic.

Step 1

Exchange Heights

↓

Step 2

Compare Tips

↓

Step 3

Compare Hashes

↓

Step 4

Request Missing Blocks

↓

Step 5

Validate

↓

Step 6

Store

↓

Step 7

Consensus

↓

Ready

---

# FORK PROTOCOL

Forks are expected.

Forks are healthy.

Forks indicate decentralization.

Fork Detection

↓

Temporary Fork

↓

Receive Competing Chain

↓

Validate Entire Chain

↓

Consensus

↓

Select Best Chain

↓

Rollback Local Chain

↓

Apply New Chain

↓

Resynchronize Wallet

↓

Continue

Wallet NEVER resolves forks.

Consensus resolves forks.

---

# CONSENSUS RESPONSIBILITIES

Consensus owns:

Chain Selection

Fork Resolution

Confirmation

Finalization

Network Agreement

Nothing else.

---

# MINING STATE MACHINE

Idle

↓

Preparing Block

↓

Loading Transactions

↓

Building Candidate

↓

Mining

↓

Found Nonce

↓

Local Validation

↓

Broadcast

↓

Consensus

↓

Reward

↓

Idle

Mining NEVER changes balances directly.

---

# REWARD PROTOCOL

Reward Proposed

↓

Consensus

↓

Accepted

↓

Added To Block

↓

Confirmed

↓

Wallet Updated

Rewards never bypass consensus.

---

# MESH MESSAGE LIFECYCLE

Message Created

↓

Assigned UUID

↓

TTL Assigned

↓

Local Queue

↓

Broadcast

↓

Neighbor Receives

↓

Duplicate Check

↓

TTL Decrement

↓

Forward

↓

Destination

↓

Acknowledgement

↓

Delivered

Expired messages are discarded.

---

# ROUTING

Every packet contains

Packet ID

Source

Destination

TTL

Timestamp

Payload Type

Payload

Checksum

No packet exists without metadata.

---

# DUPLICATE DETECTION

Packet Received

↓

Packet ID Exists?

YES

↓

Discard

NO

↓

Process

↓

Store Packet ID

Every node must remember processed packet IDs.

---

# OFFLINE MODE

When offline

Queue packets

Queue transactions

Queue blocks

Queue synchronization

Never discard valid data.

---

# RECONNECTION

Internet Available

↓

Reconnect

↓

Authenticate

↓

Exchange Heights

↓

Exchange Mempool

↓

Exchange Blocks

↓

Consensus

↓

Ready

---

# STORAGE PROTOCOL

Every write must be

Atomic

Consistent

Recoverable

Durable

Never write partial state.

---

# FAILURE RECOVERY

Crash

↓

Load Storage

↓

Verify Integrity

↓

Repair Indexes

↓

Verify Blockchain

↓

Recover Wallet

↓

Reconnect

↓

Synchronize

↓

Ready

---

# NETWORK INVARIANTS

A transaction never changes after signing.

A confirmed block never changes.

Wallet balances always come from blockchain.

Node identity never changes.

Packet IDs are unique.

Consensus is deterministic.

Serialization is deterministic.

Validation always happens before persistence.

---

# SECURITY RULES

Never trust network input.

Always validate.

Always verify signatures.

Always verify hashes.

Always verify timestamps.

Always verify packet integrity.

Reject invalid data immediately.

---

# PROTOCOL INVARIANTS

Every transaction has exactly one creator.

Every block has exactly one previous hash.

Every packet has exactly one UUID.

Every wallet has exactly one identity.

Every node has exactly one current state.

---

# AI IMPLEMENTATION RULES

Artificial Intelligence must NEVER:

Invent protocol states.

Invent synchronization shortcuts.

Invent blockchain shortcuts.

Skip validation.

Skip signatures.

Skip hashing.

Skip consensus.

Skip synchronization.

Skip recovery.

Replace deterministic behavior.

Simplify state machines.

Merge protocol states.

Remove protocol steps.

---

# GOLDEN RULE

If the implementation differs from this protocol,

THE IMPLEMENTATION IS WRONG.

The protocol is the source of truth.

END OF DOCUMENT