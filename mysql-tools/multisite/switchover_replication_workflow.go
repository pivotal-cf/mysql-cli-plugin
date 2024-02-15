package multisite

func (w Workflow) SwitchoverReplication(primaryInstance string, secondaryInstance string) error {
	w.Logger.Printf("[%s] Checking whether instance '%s' exists", w.Foundation1.ID(), primaryInstance)
	if err := w.Foundation1.InstanceExists(primaryInstance); err != nil {
		return err
	}

	w.Logger.Printf("[%s] Checking whether instance '%s' exists", w.Foundation2.ID(), secondaryInstance)
	if err := w.Foundation2.InstanceExists(secondaryInstance); err != nil {
		return err
	}

	w.Logger.Printf("[%s] Demoting primary instance '%s'", w.Foundation1.ID(), primaryInstance)
	if err := w.Foundation1.UpdateServiceAndWait(primaryInstance, `{ "initiate-failover": "make-leader-read-only" }`); err != nil {
		return err
	}

	w.Logger.Printf("[%s] Promoting secondary instance '%s'", w.Foundation2.ID(), secondaryInstance)
	if err := w.Foundation2.UpdateServiceAndWait(secondaryInstance, `{ "initiate-failover": "promote-follower-to-leader" }`); err != nil {
		return err
	}

	w.Logger.Printf("[%s] Retrieving information for new secondary instance '%s'", w.Foundation1.ID(), primaryInstance)
	hostKey, err := w.Foundation1.CreateHostInfoKey(primaryInstance)
	if err != nil {
		return err
	}

	w.Logger.Printf("[%s] Registering secondary instance information on new primary instance '%s'", w.Foundation2.ID(), secondaryInstance)
	if err = w.Foundation2.UpdateServiceAndWait(secondaryInstance, hostKey); err != nil {
		return err
	}

	w.Logger.Printf(`[%s] Retrieving replication configuration from new primary instance '%s'`, w.Foundation2.ID(), secondaryInstance)
	credKey, err := w.Foundation2.CreateCredentialsKey(secondaryInstance)
	if err != nil {
		return err
	}

	w.Logger.Printf("[%s] Updating new secondary instance '%s' with replication configuration", w.Foundation1.ID(), primaryInstance)
	if err = w.Foundation1.UpdateServiceAndWait(primaryInstance, credKey); err != nil {
		return err
	}

	w.Logger.Printf("Successfully switched replication roles. primary = [%s] %s, secondary = [%s] %s",
		w.Foundation2.ID(), secondaryInstance, w.Foundation1.ID(), primaryInstance)

	return nil
}
