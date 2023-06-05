package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Songmu/timeout"
	"github.com/viant/afs/scp"
	"golang.org/x/crypto/ssh"
)

func (b *App) readLogFromCommand(lf *LogFile) error {
	cl := strings.Split(lf.Path, " ")
	if len(cl) < 1 {
		return fmt.Errorf("command not found %s", lf.Path)
	}
	tio := &timeout.Timeout{
		Cmd:       exec.Command(cl[0], cl[1:]...),
		Duration:  120 * time.Second,
		KillAfter: 5 * time.Second,
	}
	_, stdout, _, err := tio.Run()
	if err != nil {
		return err
	}
	r := strings.NewReader(stdout)
	b.readOneLogFile(lf, r)
	return nil
}

func (b *App) readLogFromSSH(lf *LogFile) error {
	cmd := lf.Path
	if len(cmd) < 1 {
		return fmt.Errorf("command not found %s", lf.Path)
	}
	client, session, err := b.sshConnectToHost(lf)
	if err != nil {
		return err
	}
	defer func() {
		session.Close()
		client.Close()
	}()
	stdout, err := session.Output(cmd)
	if err != nil {
		return err
	}
	r := bytes.NewReader(stdout)
	b.readOneLogFile(lf, r)
	return nil
}

func (b *App) sshConnectToHost(lf *LogFile) (*ssh.Client, *ssh.Session, error) {
	kpath := lf.LogSrc.SSHKey
	if kpath == "" {
		kpath = filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa")
	}
	auth := scp.NewKeyAuth(kpath, lf.LogSrc.User, lf.LogSrc.Password)
	provider := scp.NewAuthProvider(auth, nil)
	sshConfig, err := provider.ClientConfig()
	if err != nil {
		return nil, nil, err
	}
	svr := lf.LogSrc.Server
	if !strings.Contains(svr, ":") {
		svr += ":22"
	}
	conn, err := net.DialTimeout("tcp", svr, time.Duration(60)*time.Second)
	if err != nil {
		return nil, nil, err
	}
	if err := conn.SetDeadline(time.Now().Add(time.Second * time.Duration(120))); err != nil {
		return nil, nil, err
	}
	c, ch, req, err := ssh.NewClientConn(conn, svr, sshConfig)
	if err != nil {
		return nil, nil, err
	}
	client := ssh.NewClient(c, ch, req)
	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, nil, err
	}
	return client, session, nil
}
