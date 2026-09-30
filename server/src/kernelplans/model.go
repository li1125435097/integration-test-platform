package kernelplans

import "time"

// File is the on-disk kernel-test-plans.json shape.
type File struct {
	Plans []Plan `json:"plans"`
}

// Plan is one saved kernel-test form.
type Plan struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Form      Form      `json:"form"`
}

// Form is the kernel-test page selection, shared with the browser cache.
type Form struct {
	Sources             []string `json:"sources"`
	Kernels             []string `json:"kernels"`
	Flags               []string `json:"flags"`
	CustomFlags         []string `json:"customFlags"`
	ScriptID            string   `json:"scriptId"`
	FingerprintText     string   `json:"fingerprintText"`
	FingerprintFileName string   `json:"fingerprintFileName"`
	CipherMode          bool     `json:"cipherMode"`
	TreeMode            bool     `json:"treeMode"`
	BaseURL             string   `json:"baseUrl"`
	Bearer              string   `json:"bearer"`
	ActiveSavedName     string   `json:"activeSavedName"`
	FingerprintDirty    bool     `json:"fingerprintDirty"`
}
