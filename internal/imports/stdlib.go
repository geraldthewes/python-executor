// Package imports provides automatic detection of Python package imports.
package imports

// stdlibModules contains all Python 3.12 standard library module names.
// These modules are built into Python and should not be installed via pip.
// Source: https://docs.python.org/3.12/library/index.html
var stdlibModules = map[string]bool{
	// Text Processing Services
	"string":   true,
	"re":       true,
	"difflib":  true,
	"textwrap": true,
	"unicodedata": true,
	"stringprep": true,
	"readline": true,
	"rlcompleter": true,

	// Binary Data Services
	"struct": true,
	"codecs": true,

	// Data Types
	"datetime":   true,
	"zoneinfo":   true,
	"calendar":   true,
	"collections": true,
	"heapq":      true,
	"bisect":     true,
	"array":      true,
	"weakref":    true,
	"types":      true,
	"copy":       true,
	"pprint":     true,
	"reprlib":    true,
	"enum":       true,
	"graphlib":   true,

	// Numeric and Mathematical Modules
	"numbers":   true,
	"math":      true,
	"cmath":     true,
	"decimal":   true,
	"fractions": true,
	"random":    true,
	"statistics": true,

	// Functional Programming Modules
	"itertools": true,
	"functools": true,
	"operator":  true,

	// File and Directory Access
	"pathlib":    true,
	"fileinput":  true,
	"stat":       true,
	"filecmp":    true,
	"tempfile":   true,
	"glob":       true,
	"fnmatch":    true,
	"linecache":  true,
	"shutil":     true,

	// Data Persistence
	"pickle":   true,
	"copyreg":  true,
	"shelve":   true,
	"marshal":  true,
	"dbm":      true,
	"sqlite3":  true,

	// Data Compression and Archiving
	"zlib":    true,
	"gzip":    true,
	"bz2":     true,
	"lzma":    true,
	"zipfile": true,
	"tarfile": true,

	// File Formats
	"csv":        true,
	"configparser": true,
	"tomllib":    true,
	"netrc":      true,
	"plistlib":   true,

	// Cryptographic Services
	"hashlib": true,
	"hmac":    true,
	"secrets": true,

	// Generic Operating System Services
	"os":       true,
	"io":       true,
	"time":     true,
	"argparse": true,
	"getopt":   true,
	"logging":  true,
	"getpass":  true,
	"curses":   true,
	"platform": true,
	"errno":    true,
	"ctypes":   true,

	// Concurrent Execution
	"threading":        true,
	"multiprocessing":  true,
	"concurrent":       true,
	"subprocess":       true,
	"sched":            true,
	"queue":            true,
	"contextvars":      true,

	// Networking and Interprocess Communication
	"asyncio":   true,
	"socket":    true,
	"ssl":       true,
	"select":    true,
	"selectors": true,
	"signal":    true,
	"mmap":      true,

	// Internet Data Handling
	"email":       true,
	"json":        true,
	"mailbox":     true,
	"mimetypes":   true,
	"base64":      true,
	"binascii":    true,
	"quopri":      true,

	// Structured Markup Processing Tools
	"html":        true,
	"xml":         true,

	// Internet Protocols and Support
	"webbrowser":  true,
	"wsgiref":     true,
	"urllib":      true,
	"http":        true,
	"ftplib":      true,
	"poplib":      true,
	"imaplib":     true,
	"smtplib":     true,
	"uuid":        true,
	"socketserver": true,
	"xmlrpc":      true,
	"ipaddress":   true,

	// Multimedia Services
	"wave":       true,
	"colorsys":   true,

	// Internationalization
	"gettext": true,
	"locale":  true,

	// Program Frameworks
	"turtle": true,
	"cmd":    true,
	"shlex":  true,

	// Graphical User Interfaces with Tk
	"tkinter": true,

	// Development Tools
	"typing":   true,
	"pydoc":    true,
	"doctest":  true,
	"unittest": true,
	"test":     true,

	// Debugging and Profiling
	"bdb":      true,
	"faulthandler": true,
	"pdb":      true,
	"timeit":   true,
	"trace":    true,
	"tracemalloc": true,

	// Software Packaging and Distribution
	"ensurepip":  true,
	"venv":       true,
	"zipapp":     true,

	// Python Runtime Services
	"sys":          true,
	"sysconfig":    true,
	"builtins":     true,
	"__main__":     true,
	"warnings":     true,
	"dataclasses":  true,
	"contextlib":   true,
	"abc":          true,
	"atexit":       true,
	"traceback":    true,
	"__future__":   true,
	"gc":           true,
	"inspect":      true,
	"site":         true,

	// Custom Python Interpreters
	"code":     true,
	"codeop":   true,

	// Importing Modules
	"zipimport":   true,
	"pkgutil":     true,
	"modulefinder": true,
	"runpy":       true,
	"importlib":   true,

	// Python Language Services
	"ast":       true,
	"symtable":  true,
	"token":     true,
	"keyword":   true,
	"tokenize":  true,
	"tabnanny":  true,
	"pyclbr":    true,
	"py_compile": true,
	"compileall": true,
	"dis":       true,
	"pickletools": true,

	// MS Windows Specific Services
	"msvcrt":  true,
	"winreg":  true,
	"winsound": true,

	// Unix Specific Services
	"posix":     true,
	"pwd":       true,
	"grp":       true,
	"termios":   true,
	"tty":       true,
	"pty":       true,
	"fcntl":     true,
	"resource":  true,
	"syslog":    true,

	// Superseded Modules
	"optparse": true,

	// Undocumented Modules
	"_thread": true,

	// Common submodules that should also be recognized
	"collections.abc": true,
	"os.path":         true,
	"urllib.request":  true,
	"urllib.parse":    true,
	"urllib.error":    true,
	"http.client":     true,
	"http.server":     true,
	"http.cookies":    true,
	"html.parser":     true,
	"xml.etree":       true,
	"xml.dom":         true,
	"xml.sax":         true,
	"email.mime":      true,
	"logging.handlers": true,
	"logging.config":  true,
	"unittest.mock":   true,
	"asyncio.tasks":   true,
	"asyncio.streams": true,
	"multiprocessing.pool": true,
	"concurrent.futures": true,
	"typing_extensions": true,  // Often bundled with Python
}

// IsStdlib returns true if the module name is part of the Python standard library.
func IsStdlib(module string) bool {
	return stdlibModules[module]
}
