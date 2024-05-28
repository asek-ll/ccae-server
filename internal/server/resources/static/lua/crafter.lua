local m = peripheral.find 'modem'

local config = {}

local function get_placeholder_no(name)
    local methods = peripheral.getMethods(name)
    local has_get_item_details = false
    for _, method in pairs(methods) do
        if method == 'getItemDetail' then
            local first_slot = peripheral.call(name, 'getItemDetail', 1)
            if first_slot ~= nil and first_slot['name'] == 'resourcefulbees:bee_jar' then
                return first_slot['count']
            end
        end
    end
    return nil
end

local function gen_config()
    local config = {
        inner = {},
        outer = {},
    }

    for _, name in pairs(peripheral.getNames()) do
        local no = get_placeholder_no(name)
        if no ~= nil then
            local c = config['outer']
            if name == 'top' or name == 'bottom' or name == 'front' then
                c = config['inner']
            end
            if no == 1 then
                c['input'] = name
            else
                c['output'] = name
            end
        end
    end

    if
        config['inner']['input'] == nil
        or config['inner']['output'] == nil
        or config['outer']['input'] == nil
        or config['outer']['output'] == nil
    then
        error 'wrong config'
    end

    return config
end

local function setup()
    local config = loadfile 'config.lua'
    if config == nil then
        config = gen_config()

        local f = fs.open('config.lua', 'w')
        f.write 'return '
        f.write(textutils.serialize(config))
        f.close()

        return config
    end
    return config()
end

return function(methods, handlers, wsclient)
    config = setup()
end
