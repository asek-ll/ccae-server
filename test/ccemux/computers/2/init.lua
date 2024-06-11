local m = peripheral.find 'modem' --[[@as ccTweaked.peripherals.WiredModem]]

local config = {}

local function dump_out()
    for i = 1, 16 do
        local cnt = turtle.getItemCount(i)
        if cnt > 0 then
            turtle.select(i)
            turtle.dropUp(cnt)
        end
    end
end

local function get_placeholder_no(name)
    local methods = m.getMethodsRemote(name)
    if methods == nil then
        return nil
    end
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
    local cfg = {
        buffer_side = 'top',
    }

    for _, name in pairs(m.getNamesRemote()) do
        local no = get_placeholder_no(name)
        if no ~= nil then
            print(name, no)
            if no == 1 then
                cfg['buffer_name'] = name
            end
        end
    end

    if cfg['buffer_name'] == nil then
        error 'wrong config'
    end

    return cfg
end

local function setup()
    local cfg_load = loadfile 'config.lua'
    if cfg_load == nil then
        local cfg = gen_config()

        local f = fs.open('config.lua', 'w')
        if f == nil then
            error "can't open config"
        end
        f.write 'return '
        f.write(textutils.serialize(config))
        f.close()

        return cfg
    end
    return cfg_load()
end

return function(methods, handlers, wsclient)
    config = setup()
    return config
end
