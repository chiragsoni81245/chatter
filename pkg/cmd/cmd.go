package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/chiragsoni81245/chatter/pkg/tcp"
	"github.com/spf13/cobra"
)

func clearScreen() {
    // Clearing the screen based on the operating system
    switch os_type := runtime.GOOS; os_type {
    case "darwin", "linux":
        cmd := exec.Command("clear")
        cmd.Stdout = os.Stdout
        cmd.Run()
    case "windows":
        cmd := exec.Command("cmd", "/c", "cls")
        cmd.Stdout = os.Stdout
        cmd.Run()
    default:
        fmt.Println("Unsupported platform")
    }
}


var rootCmd = &cobra.Command{
    Use: "chatter",
    Short: "A CLI chatting application",
    Long: `A CLI application where you can connect with your other fellows,
to send and receive text messages from them`,
}

func Execute(){
    err := rootCmd.Execute()
    if err != nil{
        os.Exit(1)
    }
}

var startCmd = &cobra.Command{
    Use: "start",
    Short: "start listening on given --host and --port",
    Run: startServer,
}

func init(){
    rootCmd.AddCommand(startCmd)
    startCmd.Flags().StringP("host", "a", "127.0.0.1", "Host on which you will be listening")
    startCmd.Flags().IntP("port", "p", 8080, "Port on which you will be listening")
}


type MenuOption struct {
    Name string
    ChildMenu *Menu
    Process func () error
}

type Menu struct {
    Options []*MenuOption
}

func (menu *Menu) getMenuWithPath(path []int) (*Menu, error) {
    currentMenu := menu
    for _, optionIndex := range path{
        if optionIndex >= 0 && optionIndex < len(currentMenu.Options) {
            currentMenu = currentMenu.Options[optionIndex].ChildMenu
        }else{
            return nil, errors.New("Invalid Path")
        }
    }

    return currentMenu, nil
}

func (menu *Menu) generateMenuString(currentMenuPath []int) (string, error){
    result := []string{}
    currentMenu, err := menu.getMenuWithPath(currentMenuPath)
    if err != nil {
        return "", err
    }

    for index, option := range currentMenu.Options{
        result = append(result, fmt.Sprintf("%d. %s\n", index+1, option.Name))
    }

    return strings.Join(result, ""), nil 
    
}

func (menu *Menu) processCommand(currentMenuPath []int, cmd int) error {
    currentMenu, err := menu.getMenuWithPath(currentMenuPath)
    if err != nil {
        return err
    }
    if cmd <= 0 || cmd > len(currentMenu.Options) {
        return errors.New("Invalid Command")
    }

    currentMenuPath = append(currentMenuPath, cmd-1) 
    clearScreen()

    if currentMenu.Options[cmd-1].Process == nil {
        fmt.Println("Nothing to process")
        fmt.Println("Press Enter to go back...")
        fmt.Scanln()
        return nil
    }
    err = currentMenu.Options[cmd-1].Process()
    if err != nil {
        fmt.Println(err)
        fmt.Println("Press Enter to go back...")
        fmt.Scanln()
        currentMenuPath = currentMenuPath[:len(currentMenuPath)-1]
    }
    return nil
}

func goBackWrapper(currentMenuPath []int) (func () error){
    return func () error {
        if len(currentMenuPath) == 0 {
            os.Exit(0)
        }else{
            currentMenuPath = currentMenuPath[:len(currentMenuPath)-1]
        }
        return nil
    }
}

func connectionMessagePrinter(server *tcp.TCPServer, conn *net.Conn, chatEnded chan bool, doneEmptyingCurrentQueue chan bool) {
    // This is a message reader coroutine, it will start reading messages form that particular connection's Message Queue
    var selectedConnectionAddr string = (*conn).RemoteAddr().String()
    connection := (*server).Connections[selectedConnectionAddr]
    doneEmptyingCurrentQueueFlag := false
    for {
        select {
        case <- chatEnded:
            return // Stop the message popping when the user go back from chat
        default:
            msg, err := connection.Messages.Deque()
            if err!= nil {
                if !doneEmptyingCurrentQueueFlag{
                    doneEmptyingCurrentQueue <- true
                    doneEmptyingCurrentQueueFlag = true
                }
                continue
            }
            fmt.Printf("\n\n%s> %s\n\n", selectedConnectionAddr, msg)
        }
    }
}


func sendMessageTerminal(server *tcp.TCPServer, conn *net.Conn) {
    var chatEnded chan bool = make(chan bool)
    var doneEmptyingCurrentQueue chan bool = make(chan bool)
    go connectionMessagePrinter(server, conn, chatEnded, doneEmptyingCurrentQueue) 

    <-doneEmptyingCurrentQueue // Wait for message printer to print the current messages in queue

    reader := bufio.NewReader(os.Stdin)
    for {
        fmt.Println("\nPress ESC in message to go back!")
        fmt.Print("Enter message: ")
        var msg string = "";
        msg, _ = reader.ReadString('\n')
        msg = strings.Trim(msg, "\n")
        msg = strings.TrimSpace(msg)

        if strings.Contains(msg, "\x1b") {
            fmt.Print("You pressed ESC, press enter to go back")
            fmt.Scanln()
            break
        }
        if _, ok := (*server).Connections[(*conn).RemoteAddr().String()]; !ok {
            fmt.Printf("Connection closed!, press enter to go back")
            fmt.Scanln()
            break
        }
        (*conn).Write([]byte(msg+"\n"))
        fmt.Printf("Message '%s' Sent!\n", msg)
    }

    chatEnded <- true
}

func connectWrapper(server *tcp.TCPServer) (func () error){
    return func () error{
        var address string;
        var host string;
        var port int;
        for {
            var err error;
            clearScreen()
            fmt.Print("Enter address: ")
            fmt.Scanln(&address)

            addressParts := strings.Split(address, ":")
            if len(addressParts) != 2 {
                err = errors.New("Invalid format")
            }
            host = addressParts[0] 
            if err == nil {
                port, err = strconv.Atoi(addressParts[1]) 
            }

            if err != nil {
                fmt.Printf("%s, press b to go back and any key to retry\n", err.Error())
                var input string;
                fmt.Scanln(&input)
                switch strings.TrimSpace(input) {
                case "b":
                    return nil
                default:
                    continue
                }
            }
            break
        }

        // Main Logic to connect
        connection, isConnExits := (*server).Connections[fmt.Sprintf("%s:%d", host, port)]
        var conn *net.Conn;
        var err error;
        if !isConnExits {
            conn, err = tcp.NewTCPClient(host, port)

            // This will start a receiver in background which will append the received messages into messages array of that particular connection
            (*server).AddConnection(*conn)
        }else{
            conn = &connection.Conn
        }

        if err != nil {
            fmt.Printf("Unable to connect to %s:%d\n", host, port)
            fmt.Println("Press enter to go back...")
            fmt.Scanln()
            return nil
        }
        
        
        fmt.Println("Connected!")
        
        // This will start a terminal which will take messages and send them
        sendMessageTerminal(server, conn)

        return nil
    }
}

func chatWrapper(server *tcp.TCPServer) (func () error){
    return func () error {
        var selectedConnection string;
        for {
            clearScreen()
            counter := 0
            counterToKeyMapping := make(map[int]string);

            if len((*server).Connections) == 0 {
                fmt.Println("No Connection available right now, go to connect menu to make a new connection")
                fmt.Println("or wait for someone to connect to you")
                fmt.Println("and make sure they have your listening address so that they can connect with you")
                fmt.Println()
            }

            for key := range (*server).Connections {
                counter += 1
                counterToKeyMapping[counter] = key
                messageCountString := ""
                messageCount := (*server).Connections[key].Messages.Size()
                if messageCount > 0 {
                    messageCountString = fmt.Sprintf(" (%d)", messageCount)
                }
                fmt.Printf("%d. [%s]%s\n", counter, key, messageCountString)
            }

            fmt.Println("\nType 'b' to go back")
            fmt.Print("Select a connection: ")
            // read line to get selected address 
            fmt.Scanln(&selectedConnection)

            if selectedConnection == "b"{
                return nil
            }

            selectedConnectionInt, err := strconv.Atoi(selectedConnection)
            selectedConnectionAddr, iskeyExists := counterToKeyMapping[selectedConnectionInt]

            if err!=nil || !iskeyExists {
                fmt.Println("Invalid Selection, press enter to retry")
                fmt.Scanln()
                continue
            }

            connection := (*server).Connections[selectedConnectionAddr]

            clearScreen()
            sendMessageTerminal(server, &connection.Conn)
        }
    }
}


func startServer(cmd *cobra.Command, args []string){
    host, _ := cmd.Flags().GetString("host")
    port, _ := cmd.Flags().GetInt("port")
    server, err := tcp.NewTCPServer(host, port)
    if err != nil{
        log.Fatal(err)
    }

    var currentMenuPath []int;
    menu := Menu{}
    menu.Options = []*MenuOption{
        {Name: "Connect", Process: connectWrapper(server)},
        {Name: "Chat", Process: chatWrapper(server)},
        {Name: "Exit", Process: goBackWrapper(currentMenuPath)},
    }

    for {
        clearScreen()
        headerString :=  fmt.Sprintf("Your are currently listening on %s:%d\n\n", host, port)
        menuString, err := menu.generateMenuString(currentMenuPath)
        if err != nil {
            fmt.Println(err)
            fmt.Scanln()
            continue
        }
        fmt.Println(headerString + menuString)

        var cmd int;
        fmt.Scanln(&cmd)

        menu.processCommand(currentMenuPath, cmd)
    }
}
