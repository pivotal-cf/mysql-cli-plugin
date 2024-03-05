package multisite

import (
	"fmt"
)

func (w Workflow) SwitchoverReplication(primaryInstance string, secondaryInstance string) error {
	w.Logger.Printf("[%s] Checking whether instance '%s' exists", w.Foundation1.ID(), primaryInstance)
	if err := w.Foundation1.InstanceExists(primaryInstance); err != nil {
		return err
	}

	leaderPlanName, err := w.Foundation1.InstancePlanName(primaryInstance)
	if err != nil {
		return err
	}

	w.Logger.Printf("[%s] Checking whether plan '%s' exists", w.Foundation2.ID(), leaderPlanName)
	if err = w.Foundation2.PlanExists(leaderPlanName); err != nil {
		return err
	}

	w.Logger.Printf("[%s] Checking whether instance '%s' exists", w.Foundation2.ID(), secondaryInstance)
	if err = w.Foundation2.InstanceExists(secondaryInstance); err != nil {
		return err
	}

	followerPlanName, err := w.Foundation2.InstancePlanName(secondaryInstance)
	if err != nil {
		return err
	}

	w.Logger.Printf("[%s] Checking whether plan '%s' exists", w.Foundation1.ID(), followerPlanName)
	if err = w.Foundation1.PlanExists(followerPlanName); err != nil {
		return err
	}

	w.Logger.Printf("[%s] Demoting primary instance '%s'", w.Foundation1.ID(), primaryInstance)
	if err = w.Foundation1.UpdateServiceAndWait(primaryInstance, `{ "initiate-failover": "make-leader-read-only" }`, &followerPlanName); err != nil {
		return err
	}

	w.Logger.Printf("[%s] Promoting secondary instance '%s'", w.Foundation2.ID(), secondaryInstance)
	if err = w.Foundation2.UpdateServiceAndWait(secondaryInstance, `{ "initiate-failover": "promote-follower-to-leader" }`, &leaderPlanName); err != nil {
		return err
	}

	w.Logger.Printf("[%s] Retrieving information for new secondary instance '%s'", w.Foundation1.ID(), primaryInstance)
	hostKey, err := w.Foundation1.CreateHostInfoKey(primaryInstance)
	if err != nil {
		return err
	}

	w.Logger.Printf("[%s] Registering secondary instance information on new primary instance '%s'", w.Foundation2.ID(), secondaryInstance)
	if err = w.Foundation2.UpdateServiceAndWait(secondaryInstance, hostKey, nil); err != nil {
		return err
	}

	w.Logger.Printf(`[%s] Retrieving replication configuration from new primary instance '%s'`, w.Foundation2.ID(), secondaryInstance)
	credKey, err := w.Foundation2.CreateCredentialsKey(secondaryInstance)
	if err != nil {
		return err
	}

	w.Logger.Printf("[%s] Updating new secondary instance '%s' with replication configuration", w.Foundation1.ID(), primaryInstance)
	if err = w.Foundation1.UpdateServiceAndWait(primaryInstance, credKey, nil); err != nil {
		return err
	}

	w.Logger.Printf("Successfully switched replication roles. primary = [%s] %s, secondary = [%s] %s", w.Foundation2.ID(), secondaryInstance, w.Foundation1.ID(), primaryInstance)

	return nil
}

func (w Workflow) instanceUsesPlan(foundation ServiceAPI, instance, planName string) error {
	var (
		fetchedPlanName string
		err             error
	)
	w.Logger.Printf("[%s] Checking whether plan '%s' exists", foundation.ID(), planName)
	if fetchedPlanName, err = foundation.InstancePlanName(instance); err != nil {
		return err
	}
	if fetchedPlanName != planName {
		return fmt.Errorf("foundation: %q instance: %q plan: %q does not match designated plan %q",
			foundation.ID(), instance, fetchedPlanName, planName)
	}
	return nil
}
