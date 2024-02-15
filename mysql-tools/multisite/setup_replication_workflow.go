package multisite

func (w Workflow) SetupReplication(primaryInstance string, secondaryInstance string) error {
	w.Logger.Printf("[%s] Checking whether instance '%s' exists", w.Foundation1.ID(), primaryInstance)
	if err := w.Foundation1.InstanceExists(primaryInstance); err != nil {
		return err
	}

	w.Logger.Printf("[%s] Checking whether instance '%s' exists", w.Foundation2.ID(), secondaryInstance)
	if err := w.Foundation2.InstanceExists(secondaryInstance); err != nil {
		return err
	}

	w.Logger.Printf("[%s] Retrieving information for secondary instance '%s'", w.Foundation2.ID(), secondaryInstance)
	hostKey, err := w.Foundation2.CreateHostInfoKey(secondaryInstance)
	if err != nil {
		return err
	}

	w.Logger.Printf("[%s] Registering secondary instance information on primary instance '%s'", w.Foundation1.ID(), primaryInstance)
	if err = w.Foundation1.UpdateServiceAndWait(primaryInstance, hostKey); err != nil {
		return err
	}

	w.Logger.Printf(`[%s] Retrieving replication configuration from primary instance '%s'`, w.Foundation1.ID(), primaryInstance)
	credKey, err := w.Foundation1.CreateCredentialsKey(primaryInstance)
	if err != nil {
		return err
	}

	w.Logger.Printf("[%s] Updating secondary instance '%s' with replication configuration", w.Foundation2.ID(), secondaryInstance)
	if err = w.Foundation2.UpdateServiceAndWait(secondaryInstance, credKey); err != nil {
		return err
	}

	w.Logger.Printf("Successfully configured replication")

	return nil
}
