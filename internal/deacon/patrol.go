package deacon

import (
	"fmt"
	"os/exec"

	"github.com/steveyegge/gastown/internal/beads"
	"github.com/steveyegge/gastown/internal/wisp"
)

// SlingDeaconPatrol creates a mol-deacon-patrol wisp and hooks it to the deacon.
// This is called during deacon startup to ensure the deacon has a patrol molecule
// with proper backoff, instead of improvising its own patrol loop.
//
// townRoot is the path to the town root directory.
// Returns nil if patrol was successfully slung or if deacon already has hooked work.
func SlingDeaconPatrol(townRoot string) error {
	// Check if deacon already has hooked work - skip if so
	hasWork, err := hasHookedWork(townRoot)
	if err != nil {
		return fmt.Errorf("checking for hooked work: %w", err)
	}
	if hasWork {
		// Deacon already has hooked work - don't create duplicate patrol
		return nil
	}

	// Step 1: Cook the formula (ensures proto exists)
	cookCmd := exec.Command("bd", "--no-daemon", "cook", "mol-deacon-patrol")
	cookCmd.Dir = townRoot
	if err := cookCmd.Run(); err != nil {
		return fmt.Errorf("cooking formula: %w", err)
	}

	// Step 2: Create wisp instance
	wispCmd := exec.Command("bd", "--no-daemon", "mol", "wisp", "mol-deacon-patrol", "--json")
	wispCmd.Dir = townRoot
	wispOut, err := wispCmd.Output()
	if err != nil {
		return fmt.Errorf("creating wisp: %w", err)
	}

	// Parse wisp output to get the root ID
	wispRootID, err := wisp.ParseWispIDFromJSON(wispOut)
	if err != nil {
		return fmt.Errorf("parsing wisp output: %w", err)
	}

	// Step 3: Hook the wisp to deacon
	hookCmd := exec.Command("bd", "--no-daemon", "update", wispRootID, "--status=hooked", "--assignee=deacon/")
	hookCmd.Dir = townRoot
	if err := hookCmd.Run(); err != nil {
		return fmt.Errorf("hooking wisp: %w", err)
	}

	// Step 4: Update deacon agent bead's hook_bead field using beads helper
	agentBeadID := beads.DeaconBeadIDTown()
	b := beads.New(townRoot)
	if err := b.SetHookBead(agentBeadID, wispRootID); err != nil {
		// Non-fatal: deacon can still find work via bd ready
		fmt.Printf("Warning: could not update hook_bead slot: %v\n", err)
	}

	fmt.Printf("✓ Auto-slung mol-deacon-patrol: %s\n", wispRootID)
	return nil
}

// hasHookedWork checks if the deacon already has work hooked.
// Returns true if hook_bead slot is non-empty.
func hasHookedWork(townRoot string) (bool, error) {
	agentBeadID := beads.DeaconBeadIDTown()
	b := beads.New(townRoot)

	// Get the agent bead to check hook_bead field
	_, fields, err := b.GetAgentBead(agentBeadID)
	if err != nil {
		// If bead doesn't exist yet, no hooked work
		return false, nil
	}
	if fields == nil {
		// Bead not found
		return false, nil
	}

	// Check if hook_bead field is non-empty
	return fields.HookBead != "", nil
}
