local m = peripheral.find 'modem'

local config = {}

local function get_placeholder_no(name)
    local methods = m.getMethodsRemote(name)
    print('check ' .. name)
    for _, method in pairs(methods) do
        if method == 'getItemDetail' then
            local first_slot = m.callRemote(name, 'getItemDetail', 1)
            if first_slot ~= nil and first_slot['name'] == 'resourcefulbees:bee_jar' then
                return first_slot['count']
            end
        end
    end
    return nil
end

local function gen_config()
    local config = {
        buffer_side = 'top',
    }

    for _, name in pairs(m.getNamesRemote()) do
        local no = get_placeholder_no(name)
        if no ~= nil then
            print(name, no)
            if no == 1 then
                config['buffer_name'] = name
            end
        end
    end

    if config['buffer_name'] == nil then
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
    return config
end
