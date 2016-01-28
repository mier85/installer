package installer

import (
	"github.com/kardianos/osext"

	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"text/template"
)

func GetInitDPath(name string) string {
	return "/etc/init.d/" + name
}

func IsInitDScriptExist(name string) bool {
	if _, err := os.Stat(GetInitDPath(name)); os.IsExist(err) {
		return true
	} else {
		return false
	}
}

func GetInitDScript(folder, command, user string) (ret string, err error) {

	initDTemplate, err := template.New("initd").Parse(templateInitD)
	if nil != err {
		return
	}

	buf := new(bytes.Buffer)
	initDTemplate.Execute(buf, TemplateInitD{Dir: folder, User: user, Command: folder + command, ScriptName: command})
	ret = buf.String()

	return
}

func WriteInitD(name, content string) (rErr error) {
	f, err := os.Create(GetInitDPath(name))
	if nil != err {
		rErr = err
		return
	}
	defer f.Close()
	length, err := f.Write([]byte(content))
	if nil != err {
		rErr = err
		return
	}
	if length != len(content) {
		rErr = errors.New("file could not be written completely")
		return
	}
	if err = MakeInitDExecutable(name); nil != err {
		//@TODO: also need rollback here
		rErr = err
		return
	}
	return
}

func MakeInitDExecutable(name string) (rErr error) {
	cmd := exec.Command("chmod", "+x", GetInitDPath(name))
	if _, execErr := cmd.CombinedOutput(); nil != execErr {
		rErr = execErr
	}
	return
}

func Install(user string) (rErr error) {
	pathToExe, err := osext.Executable()
	folder, command := filepath.Split(pathToExe)
	if nil != err {
		panic(err)
	}
	switch runtime.GOOS {
	case "linux":
	default:
		rErr = errors.New("not yet supported")
		return
	}
	var script string
	if IsInitDScriptExist(command) {
		rErr = errors.New(fmt.Sprintf("Script %s already exists in init.d folder", command))
		return
	}
	if script, err = GetInitDScript(folder, command, user); nil != err {
		rErr = err
		return
	}
	if fileWriteError := WriteInitD(command, script); nil != fileWriteError {
		rErr = fileWriteError
		return
	}
	/*
		cmd := exec.Command("update-rc.d", command, "defaults")
		if out, execErr := cmd.CombinedOutput(); nil != execErr {
			//@TODO: rollback , delete init.d script
			rErr = execErr
			return
		} else {
			fmt.Println(string(out))
		}
	*/
	return
}

var install = flag.Bool("install", false, "install this program as service")
var runAsUser = flag.String("installRunAsUser", "", "which user should the service run as")

func Register(parse bool) {
	if parse {
		flag.Parse()
	}
	if !*install {
		return
	}
	if *install && "" == *runAsUser {
		fmt.Println("you must specify a user that the service runs as")
		os.Exit(1)
	}
	if err := Install(*runAsUser); err != nil {
		fmt.Printf("installing as service failed: <%v> \n", err)
		os.Exit(1)
	}
	fmt.Println("successfully installed as service")
	os.Exit(0)
}

type TemplateInitD struct {
	Dir, User, Command, ScriptName string
}

const templateInitD = `#!/bin/sh
### BEGIN INIT INFO
# Provides: {{.ScriptName}}
# Required-Start: $remote_fs $syslog
# Required-Stop: $remote_fs $syslog
# Default-Start: 2 3 4 5
# Default-Stop: 0 1 6
# Short-Description: Start daemon at boot time
# Description: Enable service provided by daemon.
### END INIT INFO
dir="{{.Dir}}"
user="{{.User}}"
cmd="{{.Command}}"
name=` + "`basename $0`" + `
pid_file="/var/run/$name.pid"
stdout_log="/var/log/$name.log"
stderr_log="/var/log/$name.err"
get_pid() {
cat "$pid_file"
}
is_running() {
[ -f "$pid_file" ] && ps ` + "`get_pid`" + ` > /dev/null 2>&1
}
case "$1" in
start)
if is_running; then
echo "Already started"
else
echo "Starting $name"
cd "$dir"
sudo -u "$user" $cmd >> "$stdout_log" 2>> "$stderr_log" &
echo $! > "$pid_file"
if ! is_running; then
echo "Unable to start, see $stdout_log and $stderr_log"
exit 1
fi
fi
;;
stop)
if is_running; then
echo -n "Stopping $name.."
kill ` + "`get_pid`" + `
for i in {1..10}
do
if ! is_running; then
break
fi
echo -n "."
sleep 1
done
echo
if is_running; then
echo "Not stopped; may still be shutting down or shutdown may have failed"
exit 1
else
echo "Stopped"
if [ -f "$pid_file" ]; then
rm "$pid_file"
fi
fi
else
echo "Not running"
fi
;;
restart)
$0 stop
if is_running; then
echo "Unable to stop, will not attempt to start"
exit 1
fi
$0 start
;;
status)
if is_running; then
echo "Running"
else
echo "Stopped"
exit 1
fi
;;
*)
echo "Usage: $0 {start|stop|restart|status}"
exit 1
;;
esac
exit 0
`
