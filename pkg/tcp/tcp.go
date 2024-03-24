package tcp

import (
    "fmt"
    "log"
    "net"
    "github.com/chiragsoni81245/chatter/pkg/queue"
)

var CONNECTION_COUNTER int = 0

type Connection struct {
    Conn net.Conn
    Messages *queue.Queue
}

type TCPServer struct {
    Host string
    Port int
    Listner net.Listener
    Connections map[string]Connection
}

func (server *TCPServer) close(){
    server.Listner.Close()
}

func (server *TCPServer) AddConnection(conn net.Conn) {
    remoteAddress := conn.RemoteAddr().String()
    server.Connections[remoteAddress] = Connection{Conn: conn, Messages: &queue.Queue{}}
    go startReceiver(server, conn) 
}

func startReceiver(server *TCPServer, conn net.Conn){
    buffer := make([]byte, 1024)
    previous := ""
    for {
        n, err := conn.Read(buffer)
        if err != nil {
            conn.Close()
            delete(server.Connections, conn.RemoteAddr().String())
            break
        }

        current := previous + string(buffer[:n])
        remoteAddr := conn.RemoteAddr().String()
        if current[len(current)-1] == byte('\n') {
            if _, ok := server.Connections[remoteAddr]; ok {
                server.Connections[remoteAddr].Messages.Enque(current[:len(current)-1])
            }
            // fmt.Printf("[Conn %s][Data Received]: '%s'\n", conn.RemoteAddr().String(), current)
            current = ""
        }
    }
}

func NewTCPServer(host string, port int) (*TCPServer, error){
    listner, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))

    if err != nil {
        return nil, err
    }
    
    server := TCPServer{
        Host: host,
        Port: port,
        Listner: listner,
        Connections: make(map[string]Connection),
    }

    go func() {
        for {
            conn, err := listner.Accept()
            if err != nil {
                log.Fatal("Error in accepting connections", err)
            }
            
            server.AddConnection(conn)
        }
    }()

    return &server, err
}

func NewTCPClient(host string, port int) (*net.Conn, error) {
    conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
    if err != nil{
        return nil, err
    }

    return &conn, nil
}
