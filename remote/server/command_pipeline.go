package server

import (
	"context"

	"github.com/gomeeseeks/meeseeks-box/remote/api"
)

// CommandPipelineServer is a specific implementation for a command pipeline
type CommandPipelineServer struct {
}

// RegisterAgent registers a new agent service
func (CommandPipelineServer) RegisterAgent(in *api.AgentConfiguration, agent api.CommandPipeline_RegisterAgentServer) error {
	// When an agent is registered we need to create and add RemoteCommands to the commands map
	//
	// These commands cannot track the state as execution will happen in any order, because of this
	// they will have to contain some form of synchronization (probably a channel) which then will
	// need to be unlocked when we get the "finish" signal.
	//
	// Probably the right interface is to use an unbuffered channel that gets a
	// FinishState which will need to be managed through a map which pivots on the
	// jobID. Then the remote command will be reading from this channel such that
	// when we get the message it will unblock and return the error, if there is one.
	//
	// chan FinishState
	//
	// FinishState{
	//     Error string
	// }

	return nil
}

// Finish implements the finish server method
func (CommandPipelineServer) Finish(context.Context, *api.CommandFinish) (*api.Empty, error) {
	return nil, nil
}

// func New(address string) RemoteServer {
// 	server := grpc.NewServer()
// 	api.RegisterLogWriterServer(server, CommandLoggerServer{})
// 	api.RegisterCommandPipelineServer(server, CommandPipelineServer{})
// 	return RemoteServer{
// 		Address: address,
// 		server:  server,
// 	}
// }

// func (this RemoteServer) Listen() error {
// 	address, err := net.Listen("tcp", this.Address)
// 	if err != nil {
// 		return fmt.Errorf("could parse address %s: %s", this.Address, err)
// 	}

// 	if err := this.server.Serve(address); err != nil {
// 		return fmt.Errorf("failed to start listening on address %s: %s", this.Address, err)
// 	}
// 	return nil
// }

// // CommandLoggerServer implements the remote logger interface
// type CommandLoggerServer struct{}

// // NewAppender creates a logging stream receiver
// func (l CommandLoggerServer) NewAppender(stream api.CommandLogger_NewAppenderServer) error {
// 	for {
// 		l, err := stream.Recv()
// 		if err == io.EOF {
// 			break
// 		} else if err != nil {
// 			return err
// 		}
// 		if err := logs.Append(l.JobID, l.Line); err != nil {
// 			logrus.Errorf("Failed to record log entry %#v", l)
// 		}
// 	}
// 	return stream.SendAndClose(&api.Empty{})
// }

// // CommandPipelineServer is used to send commands to remote executors
// type CommandPipelineServer struct{}

// // RegisterAgent registers the remote agent and makes it available to start getting commands
// //
// // It receives an AgentConfiguration which declares the commands that the remote
// // executor is capable of running and a stream that will be used to send commands to
// //
// // It's not directly called, but using the remote client.
// func (c CommandPipelineServer) RegisterAgent(cfg *api.AgentConfiguration, stream api.CommandPipeline_RegisterAgentServer) error {
// 	logrus.Infof("Token: %s", cfg.Token)
// 	logrus.Infof("Labels: %s", cfg.Labels)
// 	logrus.Infof("Commands: %s", cfg.Commands)

// 	// I've a list of commands, these commands should be appended as remote commands
// 	// as a rule of thumb the way they should work is by starting a goroutine that
// 	// will wait on any command to be "executed", and when this happens, we simply
// 	// forward to command to the right downstream.
// 	//
// 	// Additionally we need to keep track of those commands so we can remove them when the
// 	// remote goes away. This should be done by token.
// 	//
// 	// This means that I need to register the remote commands in the commands map.
// 	// But then I also need to be able of removing commands from the map.

// 	var jobID uint64
// 	for {
// 		jobID++
// 		err := stream.Send(&api.CommandRequest{
// 			Command:     fmt.Sprintf("cmd-for-%s", cfg.Token),
// 			Args:        []string{"arg1", "arg2"},
// 			Channel:     "channel",
// 			ChannelID:   "channelID",
// 			ChannelLink: "channelLink",
// 			UserID:      "userID",
// 			Username:    "username",
// 			UserLink:    "userlink",
// 			JobID:       jobID,
// 			IsIM:        false,
// 		})
// 		if err == io.EOF {
// 			logrus.Info("The stream has been closed")
// 			return nil
// 		} else if err != nil {
// 			logrus.Errorf("Failed to send command %d to client: %s", jobID, err)
// 			return fmt.Errorf("something something")
// 		}
// 		select {
// 		case <-time.After(5 * time.Second):
// 			logrus.Debug("No data in over 5 seconds... looping.")
// 			continue
// 		case <-stream.Context().Done():
// 			logrus.Debug("bailing out, the context is done")
// 			return nil
// 		}
// 	}
// }

// func (c CommandPipelineServer) Finish(_ context.Context, in *api.CommandFinish) (*api.Empty, error) {
// 	logrus.Infof("Changing job %d status to %s with error %s", in.JobID, in.Status, in.Error)
// 	return &api.Empty{}, nil
// }
