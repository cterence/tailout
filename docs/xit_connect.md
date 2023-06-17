## xit connect

Connect to a machine

### Synopsis

Connect to a machine.

	This command will run tailscale up and choose the exit node with the machine name provided.
	
	Example : xit connect xit-eu-west-3-i-048afd4880f66c596

```
xit connect [machine name] [flags]
```

### Options

```
  -h, --help                help for connect
      --non-interactive     Do not prompt for confirmation
      --ts-api-key string   TailScale API Key
      --ts-tailnet string   TailScale Tailnet
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.xit.yaml)
```

### SEE ALSO

* [xit](xit.md)	 - Create an instant exit node in your tailnet

###### Auto generated by spf13/cobra on 11-Jun-2023