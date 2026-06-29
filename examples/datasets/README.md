# Benchmark Datasets

Each dataset exercises a specific optimisation behaviour.

## constructive-baseline.json

Simple dataset where constructive produces an optimal result. All items fit cleanly on their matching resources. Used as a baseline — all algorithms should produce the same score.

## skill-trap.json

A higher-priority general item (no skill requirement) is greedily assigned to the clinical specialist resource by the constructive algorithm. This blocks the specialist item that requires clinical. Hill climbing/SA should move the general item to the general resource, freeing the specialist slot.

## travel-trap.json

Constructive assigns by priority, ignoring geographic proximity. The highest-priority item is far from RES-NORTH but close to RES-SOUTH. Search algorithms may produce lower travel penalties by assigning geographically closer items to each resource.

## time-window-trap.json

The highest-priority flexible item (duration 240) is assigned first by constructive, consuming most of the shift. This may block time-window-constrained items. Search algorithms might find schedules that fit more items by respecting window ordering.

## capacity-trap.json

Multiple small items fill resources before a larger low-priority item can fit. All items have sufficient total capacity across resources but the greedy order may leave gaps. Search algorithms may rebalance to fit all items.

## preference-trap.json

Constructive assigns EVT-001 (higher priority) to RES-001 first, then EVT-002 to RES-002. But EVT-001 prefers RES-002 and EVT-002 prefers RES-001. A swap improves the preference objective without losing any assignments.
