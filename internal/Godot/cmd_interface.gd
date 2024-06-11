extends Node

var io: CmdIo

func _ready():
    io = CmdIO.new()
    io.stdin.connect(parse)

func parse(line: String) -> String:
    var line_components: Array = line.split(":", false)
    if line_components.is_empty():
        return ""
    for i in line_components.size():
        line_components[i] = str_to_var_safe(line_components[i])
    if line_components.size() >= 2:
        match line_components.pop_front():
            "call":
                return cmd_call(line_components.pop_front(), line_components.pop_front(), line_components)
            "set":
                return cmd_set(line_components.pop_front(), line_components.pop_front(), line_components.pop_front())
            "get":
                return cmd_get(line_components.pop_front(), line_components.pop_front())
    return ""

func cmd_call(object: Object, method: String, args: Array) -> String:
    if object == null:
        return ""
    print("[CALL] object: ", object, ", method: ", method, ", args: ", args)
    if object.has_method(method):
        var res: Variant = object.callv(method, args)
        if res != null:
            return var_to_str(res)
    return ""

func cmd_set(object: Object, property: String, value: Variant) -> String:
    if object == null:
        return ""
    print("[SET] object: ", object, ", property: ", property, ", value: ", value)
    if property in object:
        object.set(property, value)
        return "true"
    return "false"

func cmd_get(object: Object, property: String) -> String:
    if object == null:
        return ""
    print("[GET] object: ", object, ", property: ", property)
    return var_to_str(object.get(property))

func str_to_obj(string: String) -> Object:
    var obj: Object = null
    if string.is_valid_int():
        obj = instance_from_id(string.to_int())
    elif Engine.get_singleton_list().has(string):
        obj = Engine.get_singleton(string)
    return obj

func str_to_var_safe(string: String) -> Variant:
    var res:Variant = str_to_obj(string)
    if res == null:
        res = str_to_var(string)
    if res == null:
        return string
    return res
