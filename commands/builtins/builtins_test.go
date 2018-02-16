package builtins_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/gomeeseeks/meeseeks-box/aliases"
	"github.com/gomeeseeks/meeseeks-box/auth"
	"github.com/gomeeseeks/meeseeks-box/commands"
	"github.com/gomeeseeks/meeseeks-box/commands/builtins"
	"github.com/gomeeseeks/meeseeks-box/commands/shell"
	"github.com/gomeeseeks/meeseeks-box/jobs"
	"github.com/gomeeseeks/meeseeks-box/jobs/logs"
	"github.com/gomeeseeks/meeseeks-box/meeseeks/request"
	stubs "github.com/gomeeseeks/meeseeks-box/testingstubs"
	"github.com/gomeeseeks/meeseeks-box/tokens"
	"github.com/renstrom/dedent"
)

var basicGroups = map[string][]string{
	"admins": {"admin_user"},
	"other":  {"user_one", "user_two"},
}

var req = request.Request{
	Command:     "command",
	Channel:     "general",
	ChannelID:   "123",
	ChannelLink: "<#123>",
	UserID:      "someoneID",
	Username:    "someone",
	Args:        []string{"arg1", "arg2"},
}

func Test_BuiltinCommands(t *testing.T) {
	auth.Configure(basicGroups)

	commands.Add(builtins.BuiltinCancelJobCommand, builtins.NewCancelJobCommand(
		func(_ uint64) {}))
	commands.Add(builtins.BuiltinKillJobCommand, builtins.NewKillJobCommand(
		func(_ uint64) {}))

	tt := []struct {
		name          string
		req           request.Request
		job           jobs.Job
		setup         func()
		expected      string
		expectedMatch string
		expectedError error
	}{
		{
			name: "version command",
			req: request.Request{
				Command: builtins.BuiltinVersionCommand,
				UserID:  "userid",
			},

			job:      jobs.Job{},
			expected: "meeseeks-box version , commit , built on ",
		},
		{
			name: "help non builtins command",
			req: request.Request{
				Command: builtins.BuiltinHelpCommand,
				UserID:  "userid",
			},

			job:      jobs.Job{Request: request.Request{Args: []string{}}},
			expected: "",
		},
		{
			name: "help all command",
			req: request.Request{
				Command: builtins.BuiltinHelpCommand,
				UserID:  "userid",
			},
			job: jobs.Job{Request: request.Request{Args: []string{"-all"}}},
			expected: dedent.Dedent(`- alias: adds an alias for a command for the current user
				- aliases: list all the aliases for the current user
				- audit: lists jobs from all users or a specific one (admin only)
				- auditjob: shows a command metadata by job ID (admin only)
				- auditlogs: shows the logs of a job by ID (admin only)
				- cancel: sends a cancellation signal to a job owned by the current user
				- groups: prints the configured groups
				- help: shows the help for all the commands, or a single one
				- job: show metadata of one job by id
				- jobs: shows the last executed jobs for the calling user
				- kill: sends a cancellation signal to a job, admin only
				- last: shows the last job metadata executed by the current user
				- logs: returns the full output of the job passed as argument
				- tail: returns the last lines of the last executed job, or one selected by job ID
				- token-new: creates a new API token
				- token-revoke: revokes an API token
				- tokens: lists the API tokens
				- unalias: deletes an alias
				- version: prints the running meeseeks version
				`),
		},
		{
			name: "help one command",
			req: request.Request{
				Command: builtins.BuiltinHelpCommand,
				UserID:  "userid",
			},
			job: jobs.Job{Request: request.Request{Args: []string{"token-new"}}},
			expected: dedent.Dedent(`*token-new* - creates a new API token
				
				*Arguments*
				- user that will be impersonated by the api, mandatory
				- channel that will be used as the one in which the job was called
				- command the token will be calling
				- arguments to pass to the command
				`),
		},
		{
			name: "groups command",
			req: request.Request{
				Command: builtins.BuiltinGroupsCommand,
				UserID:  "userid",
			},

			job: jobs.Job{},
			expected: dedent.Dedent(`
					- admins: admin_user
					- other: user_one, user_two
					`),
		},
		{
			name: "test jobs command",
			req: request.Request{
				Command: builtins.BuiltinJobsCommand,
				UserID:  "userid",
			},

			job: jobs.Job{
				Request: request.Request{Username: "someone"},
			},
			setup: func() {
				j, err := jobs.Create(req)
				stubs.Must(t, "could not create job", err)
				j.Finish(jobs.SuccessStatus)
			},
			expected: "*1* - now - *command* by *someone* in *<#123>* - *Successful*\n",
		},
		{
			name: "test audit command",
			req: request.Request{
				Command: builtins.BuiltinAuditCommand,
				UserID:  "userid",
			},

			job: jobs.Job{
				Request: request.Request{},
			},
			setup: func() {
				j, err := jobs.Create(req)
				stubs.Must(t, "could not create job", err)
				j.Finish(jobs.SuccessStatus)
			},
			expected: "*1* - now - *command* by *someone* in *<#123>* - *Successful*\n",
		},
		{
			name: "test jobs command with limit",
			req: request.Request{
				Command: builtins.BuiltinJobsCommand,
				UserID:  "userid",
			},

			job: jobs.Job{
				Request: request.Request{Username: "someone", Args: []string{"-limit=1"}},
			},
			setup: func() {
				jobs.Create(req)
				jobs.Create(req)
			},
			expected: "*2* - now - *command* by *someone* in *<#123>* - *Running*\n",
		},
		{
			name: "test jobs command on IM",
			req: request.Request{
				Command: builtins.BuiltinJobsCommand,
				UserID:  "userid",
			},

			job: jobs.Job{
				Request: request.Request{Username: "someone"},
			},
			setup: func() {
				jobs.Create(request.Request{
					Command:   "command",
					Channel:   "general",
					ChannelID: "123",
					Username:  "someone",
					Args:      []string{"arg1", "arg2"},
					IsIM:      true,
				})
			},
			expected: "*1* - now - *command* by *someone* in *DM* - *Running*\n",
		},
		{
			name: "test last command",
			req: request.Request{
				Command: builtins.BuiltinLastCommand,
				UserID:  "userid",
			},

			job: jobs.Job{
				Request: request.Request{Username: "someone"},
			},
			setup: func() {
				jobs.Create(req)
				jobs.Create(req)
				jobs.Create(req)
			},
			expected: "* *ID* 3\n* *Status* Running\n* *Command* command\n* *Args* \"arg1\" \"arg2\" \n* *Where* <#123>\n* *When* now\n",
		},
		{
			name: "test find command",
			req: request.Request{
				Command: builtins.BuiltinFindJobCommand,
				UserID:  "userid",
			},

			job: jobs.Job{
				Request: request.Request{Username: "someone", Args: []string{"1"}},
			},
			setup: func() {
				jobs.Create(req)
				jobs.Create(req)
			},
			expected: "* *ID* 1\n* *Status* Running\n* *Command* command\n* *Args* \"arg1\" \"arg2\" \n* *Where* <#123>\n* *When* now\n",
		},
		{
			name: "test alias command",
			req: request.Request{
				Command: builtins.BuiltinNewAliasCommand,
				UserID:  "userid",
			},

			job: jobs.Job{
				Request: request.Request{
					Command: "alias",
					Args:    []string{"command", "for", "-add", "alias"},
					UserID:  "userid",
				}},
			expected: "alias created successfully",
		},
		{
			name: "test aliases command",
			req: request.Request{
				Command: builtins.BuiltinGetAliasesCommand,
				UserID:  "userid",
			},

			job: jobs.Job{
				Request: request.Request{
					Command: "aliases",
					UserID:  "userid",
				},
			},
			setup: func() {
				err := aliases.Create("userid", "first", "command", []string{"-with args"}...)
				stubs.Must(t, "create first alias", err)

				err = aliases.Create("userid", "second", "another", []string{"-command"}...)
				stubs.Must(t, "create second alias", err)
			},
			expected: "- *first* - `command -with args`\n- *second* - `another -command`\n",
		},
		{
			name: "test unalias command",
			req: request.Request{
				Command: builtins.BuiltinGetAliasesCommand,
				UserID:  "userid",
			},

			job: jobs.Job{
				Request: request.Request{
					Command: "aliases",
					UserID:  "userid",
				},
			},
			setup: func() {
				err := aliases.Create("userid", "command", "command", []string{"-with", "args"}...)
				stubs.Must(t, "create first alias", err)

				err = aliases.Create("userid", "second", "another", []string{"-command"}...)
				stubs.Must(t, "create second alias", err)

				err = aliases.Delete("userid", "command")
				stubs.Must(t, "delete first alias", err)

			},
			expected: "- *second* - `another -command`\n",
		},
		{
			name: "test alias execution",
			req: request.Request{
				Command: "testalias",
				UserID:  "userid",
			},

			job: jobs.Job{
				Request: request.Request{
					Command: "testalias",
					UserID:  "userid",
				},
			},
			setup: func() {
				err := aliases.Create("userid", "testalias", "audit", []string{"-limit", "1"}...)
				stubs.Must(t, "create an alias", err)

				commands.Add("noop", shell.New(shell.CommandOpts{
					AuthStrategy: "any",
					Cmd:          "true",
				}))

				_, err = jobs.Create(
					request.Request{
						Command: "noop",
						UserID:  "userid",
					})
				stubs.Must(t, "do nothing", err)

			},
			expected: "*1* - now - *noop* by ** in ** - *Running*\n",
		},
		{
			name: "test auditjob command",
			req: request.Request{
				Command: builtins.BuiltinAuditJobCommand,
				UserID:  "userid",
			},

			job: jobs.Job{
				Request: request.Request{Username: "someone", Args: []string{"1"}},
			},
			setup: func() {
				jobs.Create(req)
				jobs.Create(req)
			},
			expected: "* *ID* 1\n* *Status* Running\n* *Command* command\n* *Args* \"arg1\" \"arg2\" \n* *Where* <#123>\n* *When* now\n",
		},
		{
			name: "test tail command",
			req: request.Request{
				Command: builtins.BuiltinTailCommand,
				UserID:  "userid",
			},

			job: jobs.Job{
				Request: request.Request{Username: "someone"},
			},
			setup: func() {
				j, err := jobs.Create(req)
				stubs.Must(t, "create job", err)
				logs.Append(j.ID, "something to say 1")

				j, err = jobs.Create(req)
				stubs.Must(t, "create job", err)
				logs.Append(j.ID, "something to say 2")
			},
			expected: "something to say 2",
		},
		{
			name: "test logs command",
			req: request.Request{
				Command: builtins.BuiltinLogsCommand,
				UserID:  "userid",
			},

			job: jobs.Job{
				Request: request.Request{Username: "someone", Args: []string{"1"}},
			},
			setup: func() {
				j, err := jobs.Create(req)
				stubs.Must(t, "create job", err)
				logs.Append(j.ID, "something to say 1")

				j, err = jobs.Create(req)
				stubs.Must(t, "create job", err)
				logs.Append(j.ID, "something to say 2")
			},
			expected: "something to say 1",
		},
		{
			name: "test auditlogs command",
			req: request.Request{
				Command: builtins.BuiltinAuditLogsCommand,
				UserID:  "userid",
			},

			job: jobs.Job{
				Request: request.Request{Username: "admin_user", Args: []string{"1"}},
			},
			setup: func() {
				j, err := jobs.Create(req)
				stubs.Must(t, "create job", err)
				logs.Append(j.ID, "something to say 1")

				j, err = jobs.Create(req)
				stubs.Must(t, "create job", err)
				logs.Append(j.ID, "something to say 2")
			},
			expected: "something to say 1",
		},
		{
			name: "test token-new command",
			req: request.Request{
				Command: builtins.BuiltinNewAPITokenCommand,
				UserID:  "userid",
			},

			job: jobs.Job{
				Request: request.Request{Username: "admin_user", IsIM: true, Args: []string{"admin_user", "yolo", "rm", "-rf"}},
			},
			expectedMatch: "created token .*",
		},
		{
			name: "test tokens command",
			req: request.Request{
				Command: builtins.BuiltinListAPITokenCommand,
				UserID:  "userid",
			},

			job: jobs.Job{
				Request: request.Request{Username: "admin_user", IsIM: true},
			},
			setup: func() {
				_, err := tokens.Create(tokens.NewTokenRequest{
					ChannelLink: "channelLink",
					UserLink:    "userLink",
					Text:        "something",
				})
				stubs.Must(t, "create token", err)

			},
			expectedMatch: "- \\*.*?\\* userLink at channelLink _something_",
		},
		{
			name: "test kill job command",
			req: request.Request{
				Command: builtins.BuiltinKillJobCommand,
				UserID:  "userid",
			},

			job: jobs.Job{
				Request: request.Request{Username: "someone", Args: []string{"1"}},
			},
			setup: func() {
				_, err := jobs.Create(req)
				stubs.Must(t, "create job", err)

				_, err = jobs.Create(req)
				stubs.Must(t, "create job", err)
			},
			expected: "Issued command cancellation to job 1",
		},
		{
			name: "test cancel job command",
			req: request.Request{
				Command: builtins.BuiltinCancelJobCommand,
				UserID:  "userid",
			},

			job: jobs.Job{
				Request: request.Request{Username: "someone", Args: []string{"2"}},
			},
			setup: func() {
				_, err := jobs.Create(req)
				stubs.Must(t, "create job", err)

				_, err = jobs.Create(req)
				stubs.Must(t, "create job", err)
			},
			expected: "Issued command cancellation to job 2",
		},
		{
			name: "test cancel job command with wrong user",
			req: request.Request{
				Command: builtins.BuiltinCancelJobCommand,
				UserID:  "userid",
			},

			job: jobs.Job{
				Request: request.Request{Username: "someone_else", Args: []string{"2"}},
			},
			setup: func() {
				_, err := jobs.Create(req)
				stubs.Must(t, "create job", err)

				_, err = jobs.Create(req)
				stubs.Must(t, "create job", err)
			},
			expectedError: fmt.Errorf("no job could be found"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			stubs.Must(t, "failed to run tests", stubs.WithTmpDB(func(_ string) {
				if tc.setup != nil {
					tc.setup()
				}
				cmd, ok := commands.Find(&tc.req)
				if !ok {
					t.Fatalf("could not find command %s", tc.req.Command)
				}

				out, err := cmd.Execute(context.Background(), tc.job)
				if err != nil && tc.expectedError != nil {
					stubs.AssertEquals(t, err.Error(), tc.expectedError.Error())
					return
				}

				stubs.Must(t, "cmd erred out", err)
				if tc.expected != "" {
					stubs.AssertEquals(t, tc.expected, out)
				}
				if tc.expectedMatch != "" {
					stubs.AssertMatches(t, tc.expectedMatch, out)
				}
			}))
		})
	}
}

func Test_FilterJobsAudit(t *testing.T) {
	stubs.Must(t, "failed to audit the correct jobs", stubs.WithTmpDB(func(_ string) {
		r1 := request.Request{
			Command:     "command",
			Channel:     "general",
			ChannelID:   "123",
			ChannelLink: "<#123>",
			Username:    "someone",
			Args:        []string{"some", "thing"},
		}
		r2 := request.Request{
			Command:     "command",
			Channel:     "general",
			ChannelID:   "123",
			ChannelLink: "<#123>",
			Username:    "someoneelse",
			Args:        []string{"something", "else"},
		}

		auth.Configure(basicGroups)
		jobs.Create(r1)
		jobs.Create(r2)
		jobs.Create(r1)
		jobs.Create(r1)
		jobs.Create(r2)

		cmd, ok := commands.Find(&request.Request{
			Command: "audit",
			UserID:  "userid",
		})
		if !ok {
			t.Fatalf("could not find command %s", "audit")
		}

		audit, err := cmd.Execute(context.Background(), jobs.Job{
			Request: request.Request{Args: []string{"-user", "someone"}},
		})
		if err != nil {
			t.Fatalf("Failed to execute audit: %s", err)
		}
		stubs.AssertEquals(t, "*4* - now - *command* by *someone* in *<#123>* - *Running*\n*3* - now - *command* by *someone* in *<#123>* - *Running*\n*1* - now - *command* by *someone* in *<#123>* - *Running*\n", audit)

		limit, err := cmd.Execute(context.Background(), jobs.Job{
			Request: request.Request{Args: []string{"-user", "someone", "-limit", "2"}},
		})
		if err != nil {
			t.Fatalf("Failed to execute audit: %s", err)
		}
		stubs.AssertEquals(t, "*4* - now - *command* by *someone* in *<#123>* - *Running*\n*3* - now - *command* by *someone* in *<#123>* - *Running*\n", limit)
	}))
}

func TestAPITokenLifecycle(t *testing.T) {

	exec := func(r request.Request) (string, error) {
		cmd, ok := commands.Find(&r)
		if !ok {
			t.Fatalf("could not find command %s", r.Command)
		}

		return cmd.Execute(context.Background(), jobs.NullJob(r))
	}

	stubs.Must(t, "failed to audit the correct jobs", stubs.WithTmpDB(func(_ string) {

		out, err := exec(request.Request{
			Command: builtins.BuiltinListAPITokenCommand,
			UserID:  "apiuser",
			IsIM:    true,
		})
		stubs.Must(t, "can't list api tokens:", err)
		stubs.AssertEquals(t, "No tokens could be found", out)

		out, err = exec(request.Request{
			Command: builtins.BuiltinNewAPITokenCommand,
			UserID:  "apiuser",
			IsIM:    true,
			Args:    []string{"apiuser", "yolo", "rm", "-rf"},
		})
		stubs.Must(t, "can't create an api token:", err)

		token := strings.Split(out, " ")[2]

		out, err = exec(request.Request{
			Command: builtins.BuiltinListAPITokenCommand,
			UserID:  "apiuser",
			IsIM:    true,
		})
		stubs.Must(t, "can't list api tokens:", err)
		stubs.AssertEquals(t, fmt.Sprintf("- *%s* apiuser at yolo _rm -rf_\n", token), out)

		out, err = exec(request.Request{
			Command: builtins.BuiltinRevokeAPITokenCommand,
			UserID:  "apiuser",
			IsIM:    true,
			Args:    []string{token},
		})
		stubs.Must(t, "can't revoke an api token:", err)
		stubs.AssertEquals(t, fmt.Sprintf("Token *%s* has been revoked", token), out)

		out, err = exec(request.Request{
			Command: builtins.BuiltinListAPITokenCommand,
			UserID:  "apiuser",
			IsIM:    true,
		})
		stubs.Must(t, "can't list api tokens:", err)
		stubs.AssertEquals(t, "No tokens could be found", out)
	}))
}
