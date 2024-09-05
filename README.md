# `fakemoderun` - Affinitize Cores to Processes

`fakemoderun` is a simple cli application which can be used to affintiize cores to
processes from the command line in windows . The aim is to be similar to Linux's
`taskset` command. 

## Usage

```powershell
fakemoderun.exe -cores <core range> <command> <args>
```

e.g.

```powershell
fakemoderun.exe -cores 1-8 notepad.exe test.txt

```

You can use this with steam in the same way as you'd use a program like
`gamemoderun` i.e. by using it in your launch options.

```
fakemoderun.exe -cores 0-7,16-23 %command%
```

## Caveats

- `fakemoderun` only affinitizes cores to a speciifc process. It does not evict
existing processes from those cores and does not park any cores.

- Processes may try to affintize themselves after being affinitized by `fakemoderun`.

- This application was written for use only on Windows systems. Linux users
should use the far superior `taskset` or `gamemoderun` commands.
