class_name CmdIO extends RefCounted

signal stdin(line: String)

# Internal state
var _stdin_thread: Thread
var _stdin_running: bool

func write(stdout: String) -> void:
    print(stdout)

func _init(read_stdin: bool = false) -> void:
    if read_stdin:
        _stdin_running = true
        _stdin_thread = Thread.new()
        _stdin_thread.start(_read_stdin)

func _read_stdin() -> void:
    var stdin: String
    while _stdin_running:
        stdin = OS.read_string_from_stdin()
        if stdin != "":
            _process_stdin.call_deferred(stdin)

func _on_close_stdin() -> void:
    if _stdin_thread != null and !_stdin_running and _stdin_thread.is_started() and !_stdin_thread.is_alive():
        _stdin_thread.wait_to_finish()

func _process_stdin(value: String) -> void:
    stdin.emit(value)
