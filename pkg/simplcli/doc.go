// Package simplcli provides a simple CLI (command line interface) implementation.
// It is a lightweight and easy-to-use library for building subcommands and applying middleware to CLI applications.
// It uses no dependencies outside of the Go stdlib.
//
// A [SimplCLI] is a collection of sub commands, which are just
// a function matching the [Runner] interface and a doc string.
//
//	greetRunner := func(ctx context.Context, stdout, stderr io.Writer, args []string) error {
//		fmt.Fprintln(stdout, "hello!")
//		return nil
//	}
//
//	cli := simplcli.SimplCLI{
//		SubCmds: map[string]simplcli.SubCmd{
//			"hello": {Runner: greetRunner, Doc: "say hello"},
//		},
//	}
//
//	simplcli.Entrypoint(cli)
//
// # Nesting Subcommands
//
// [SimplCLI] can be nested by using the [SimplCLI.Run] as the [Runner]
// in another [SimplCLI]'s sub commands:
//
//	greetRunner := func(ctx context.Context, stdout, stderr io.Writer, args []string) error {
//		fmt.Fprintln(stdout, "hello, world!")
//		return nil
//	}
//
//	childCLI := simplcli.SimplCLI{
//		SubCmds: map[string]simplcli.SubCmd{
//			"greet": {Runner: greetRunner, Doc: "greet the world"},
//		},
//	}
//
//	parentCLI := simplcli.SimplCLI{
//		SubCmds: map[string]simplcli.SubCmd{
//			"sub": {Runner: childCLI.Run, Doc: "a nested subcommand"},
//		},
//	}
//	err := parentCLI.Run(ctx, os.Stdout, os.Stderr, []string{"sub", "greet"})
//
// # Middleware
//
// To facilitate setup/teardown (e.g. database connection, logging,
// etc.pp) [Middleware] can be used. Middlewares set on a [SimplCLI]
// are executed in order for every sub command.
//
//	logMiddleware := func(ctx context.Context, stdout, stderr io.Writer, args []string, next Runner) error {
//		fmt.Fprintln(stdout, "before command")
//		err := next(ctx, stdout, stderr, args)
//		fmt.Fprintln(stdout, "after command")
//		return err
//	}
//
//	cliWithMiddleware := simplcli.SimplCLI{
//		Middlewares: []simplcli.Middleware{logMiddleware},
//		SubCmds: map[string]simplcli.SubCmd{
//			"greet": {Runner: greetRunner, Doc: "greet someone"},
//		},
//	}
//	err := cliWithMiddleware.Run(ctx, os.Stdout, os.Stderr, []string{"greet"})
//
// # Flag handling
//
// Handling flags is not easy - the stdlib flag package is fantastic but it also has drawbacks.
//
// Two approaches that work well with simplcli is either using package
// level flags and calling [flag.Parse] in a middleware or using [flag.FlagSet].
//
// flag.Parse with package level flags works well for small-ish cli
// applications that have many shared global values.
//
//	var myFlag = flag.String("flag", "default", "docs...")
//
//	cli := simplcli.SimplCLI{
//		Middlewares: []simplcli.Middleware{
//			func(ctx context.Context, stdout, stderr io.Writer, args []string, next Runner) error {
//				flag.Parse()
//				return next(ctx, stdout, stderr, flag.Args())
//			},
//		},
//		SubCmds: map[string]simplcli.SubCmd{
//			// ...
//		},
//	}
//
//	simplcli.Entrypoint(cli)
//
// flag.FlagSet has the same features as the package level flags but
// allows building the flagset for each sub command specifically.
//
// A common pattern is to have a method RegisterFlags on structs,
// which take a [flag.FlagSet] and register their flags on it.
//
//	type Options struct {
//		SomeAttribute string
//	}
//
//	func (g *Greeter) RegisterFlags(fs *flags.FlagSet) {
//		fs.StringVar(&o.SomeAttribute, "", false, "Attribute content")
//	}
//
//	greetRunner := func(ctx context.Context, stdout, stderr io.Writer, args []string) error {
//		fs := flag.NewFlagSet("", flag.ExitOnError)
//		fs.SetOutput(stderr)
//		fName := fs.String("name", "", "Name to greet")
//
//		o := &Options{}
//		o.RegisterFlags(fs)
//
//		if err := fs.Parse(args); err != nil {
//			return err
//		}
//		_, err := fmt.Fprintf(stdout, "hello %s!", *fName)
//		return err
//	}
//
//	cli := simplcli.SimplCLI{
//		SubCmds: map[string]simplcli.SubCmd{
//			"greet": {Runner: greetRunner, Doc: "greet someone"},
//		},
//	}
//	simplcli.Entrypoint(cli)
package simplcli
