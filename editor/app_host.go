package main

import (
        "fmt"
)



func (a *App) HostGame(projectPath string, port int) (string, error) {
        return a.BuildPublicPack(projectPath)
}


func (a *App) StopHost() error {
        a.hostMu.Lock()
        defer a.hostMu.Unlock()
        if a.hostServer != nil {
                err := a.hostServer.Close()
                a.hostServer = nil
                a.hostPort = 0
                return err
        }
        return nil
}


func (a *App) GetHostStatus() (bool, int, string) {
        a.hostMu.Lock()
        defer a.hostMu.Unlock()
        if a.hostServer != nil && a.hostPort != 0 {
                return true, a.hostPort, fmt.Sprintf("http://localhost:%d", a.hostPort)
        }
        return false, 0, ""
}

