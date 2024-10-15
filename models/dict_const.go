package models

type VRUrgency string

const (
	VRTypeUrgent    VRUrgency = "срочный"
	VRTypeNonUrgent VRUrgency = "несрочный"
)

type VRType string

const (
	VRTypeMass       VRType = "массовый подбор"
	VRTypeCommercial VRType = "коммерческий"
	VRTypePersonal   VRType = "персональный"
)
