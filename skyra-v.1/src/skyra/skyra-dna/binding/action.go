package binding

// Action — execution tracking.
// Indexes by primitive → instance → execution result.
// Records whether the motor command executed, not the world's response.
// Deposit: every time motor fires and returns.
// Proposer: collapses to the primitive that executed successfully before.
