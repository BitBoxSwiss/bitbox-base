package rpcmessages

/*
Put notification constants here. Notifications for new rpc data should have the format 'OpUCanHas' + 'RPC Method Name'.
*/
const (
	// OpRPCCall is prepended to every rpc response messages, to indicate that the message is rpc response and not a notification.
	OpRPCCall = "r"
	// OpUCanHasSampleInfo notifies when new SampleInfo data is available.
	OpUCanHasSampleInfo = "d"
	// OpUCanHasVerificationProgress notifies when new VerificationProgress data is available.
	OpUCanHasVerificationProgress = "v"
)

/*
Put Incoming Args below this line. They should have the format of 'RPC Method Name' + 'Args'.
*/

// ResyncBitcoinArgs is an iota that holds the options for the ResyncBitcoin rpc call
type ResyncBitcoinArgs int

// The ResyncBitcoinArgs has two options. Other resync bitcon from scratch with an IBD, or delete the chainstate and reindex.
const (
	Resync ResyncBitcoinArgs = iota
	Reindex
)

// FlashdriveArgs is an struct that holds the arguments for the Flashdrive RPC call
type FlashdriveArgs struct {
	Method FlashdriveMethod // the method called
	Path   string           // the method 'mount' needs a path. If not calling 'mount' this path should be empty.
}

// FlashdriveMethod is an iota that holds the method for the Flashdrive RPC call
type FlashdriveMethod int

// FlashdriveMethod can be one of three possible methods.
// Either check for an existing flashdrive, mount a flash drive or unmount a mounted drive.
const (
	Check FlashdriveMethod = iota
	Mount
	Unmount
)

// BackupArgs is an iota that holds the method for the Backup RPC call
type BackupArgs int

// The BackupArgs has two methods. Backup the system config (sysconfig) or the hsm_secret by c-lightning
const (
	SysConfig BackupArgs = iota
	HSMSecret
)

/*
Put Response structs below this line. They should have the format of 'RPC Method Name' + 'Response'.
*/

// GetEnvResponse is the struct that gets sent by the rpc server during a GetSystemEnv call
type GetEnvResponse struct {
	Network        string
	ElectrsRPCPort string
}

// ResyncBitcoinResponse is the struct that gets sent by the rpc server during a ResyncBitcoin call
type ResyncBitcoinResponse struct {
	Success bool
}

// SampleInfoResponse holds sample information from c-lightning and bitcoind. It is temporary for testing purposes
type SampleInfoResponse struct {
	Blocks         int64   `json:"blocks"`
	Difficulty     float64 `json:"difficulty"`
	LightningAlias string  `json:"lightningAlias"`
}

// VerificationProgressResponse is the struct that gets sent by the rpc server during a VerificationProgress rpc call
type VerificationProgressResponse struct {
	Blocks               int64   `json:"blocks"`
	Headers              int64   `json:"headers"`
	VerificationProgress float64 `json:"verificationProgress"`
}

// FlashdriveResponse is the struct that gets sent by the rpc server during a Flashdrive call
type FlashdriveResponse struct {
	Success bool
	Message string
}

// BackupResponse is the struct that gets sent by the rpc server during a Backup call
type BackupResponse struct {
	Success bool
	Message string
}
