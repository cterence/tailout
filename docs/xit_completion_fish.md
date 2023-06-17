## xit completion fish

Generate the autocompletion script for fish

### Synopsis

Generate the autocompletion script for the fish shell.

To load completions in your current shell session:

	xit completion fish | source

To load completions for every new session, execute once:

	xit completion fish > ~/.config/fish/completions/xit.fish

You will need to start a new shell for this setup to take effect.


```
xit completion fish [flags]
```

### Options

```
  -h, --help              help for fish
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.xit.yaml)
```

### SEE ALSO

* [xit completion](xit_completion.md)	 - Generate the autocompletion script for the specified shell

###### Auto generated by spf13/cobra on 11-Jun-2023