package env

import "embed"

var Files embed.FS

const VERSION = "v0.0.1"
const BANNER = ""

var UNIX = false
var UseEnvFile = false
var Host string
var NameSpace string
var Testing = false
var Port = 80
var TimeAlive = 5
var SimultaneousInstances = 10

var CheckTime = false
var CheckSimultaneous = false
