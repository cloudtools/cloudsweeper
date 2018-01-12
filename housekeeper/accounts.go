package housekeeper

// Owner contains an AWS account ID and the name of the owner.
// This name is typically the chosen brkt username, and can be
// used for emailing.
type Owner struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

// Owners is a collection of Owner, with some associated
// helper methods
type Owners []Owner

// AllIDs return only a list of IDs for all owners
func (o *Owners) AllIDs() []string {
	ids := []string{}
	for i := range *o {
		ids = append(ids, (*o)[i].ID)
	}
	return ids
}

// NameToID create a mapping from owner IDs to
// owner Names
func (o *Owners) NameToID() map[string]string {
	result := make(map[string]string)
	for i := range *o {
		result[(*o)[i].Name] = (*o)[i].ID
	}
	return result
}

// IDToName create a mapping from owner Names to
// owner IDs
func (o *Owners) IDToName() map[string]string {
	result := make(map[string]string)
	for i := range *o {
		result[(*o)[i].ID] = (*o)[i].Name
	}
	return result
}
