local m = peripheral.find 'modem' --[[@as ccTweaked.peripherals.WiredModem]]
local source = peripheral.wrap 'top' --[[@as ccTweaked.peripherals.Inventory]]
-- local buf = peripheral.wrap 'bottom' --[[@as ccTweaked.peripherals.Inventory]]

local config = {}

local function dump_out()
    for i = 1, 16 do
        local cnt = turtle.getItemCount(i)
        if cnt > 0 then
            turtle.select(i)
            turtle.dropUp(cnt)
        end
    end
    source.pullItems('bottom', 1)
    return true
end

local function craft_slot_to_turtle_slot(slot)
    local column = (slot - 1) % 3
    local row = (slot - 1 - column) / 3
    local result_slot = (row * 4 + column) + 1
    return result_slot
end

local function craft()
    local items = source.list()
    for slot = 1, 9 do
        if items[slot] ~= nil then
            source.pushItems('bottom', slot)
            turtle.select(craft_slot_to_turtle_slot(slot))
            turtle.suckDown()
        end
    end
    turtle.select(13)
    local res, msg = turtle.craft()
    if res then
        turtle.select(13)
        turtle.drop()
    end
    return res
end

local function restore()
    if turtle.getItemCount(13) > 0 then
        turtle.select(13)
        turtle.drop()
        return true
    end
    return craft()
end

local function processResults(storage)
    for slot in pairs(m.callRemote(config["output_name"], "list")) do
        m.callRemote(config['output_name'], "pushItems", storage, slot)
    end
    return true
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
    methods['dumpOut'] = dump_out
    methods['craft'] = craft
    methods['restore'] = restore
    methods['processResults'] = processResults
    return config
end
