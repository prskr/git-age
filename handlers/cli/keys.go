package cli

type KeysCliHandler struct {
	Generate GenKeyCliHandler   `cmd:"" name:"generate" aliases:"gen" help:"Generate a new key pair"`
	List     ListKeysCliHandler `cmd:"" name:"list" aliases:"ls" help:"List all keys"`
}
