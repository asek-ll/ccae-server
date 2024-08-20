local m = peripheral.find 'modem' --[[@as ccTweaked.peripherals.WiredModem]]

local config = {}

---@param modem ccTweaked.peripherals.WiredModem
---@param storage_name string
---@return boolean
local function is_storage(modem, storage_name)
    local methods = modem.getMethodsRemote(storage_name)
    if methods == nil then
        return false
    end
    for _, method in pairs(methods) do
        if method == 'getItemDetail' then
            return true
        end
    end
    return false
end

local function json_list(table)
    if #table == 0 then
        return textutils.empty_json_array
    end
    return table
end

local function get_items(storage_prefixes)
    local storages = {}

    for _, storage_name in pairs(m.getNamesRemote()) do
        local valid = false
        for _, prefix in pairs(storage_prefixes) do
            if string.sub(storage_name, 1, string.len(prefix)) == prefix then
                valid = true
                break
            end
        end
        if valid and is_storage(m, storage_name) then
            local size = m.callRemote(storage_name, 'size')

            local storage_items = {}
            local items = m.callRemote(storage_name, 'list')
            for slot, item in pairs(items) do
                -- local details = m.callRemote(storage_name, 'getItemDetail', slot)

                local details = {
                    name = item['name'],
                    nbt = item['nbt'],
                    count = item['count'],
                    maxCount = item['count'],
                }

                if details ~= nil then
                    local cache_item = { item = details, slot = slot }
                    table.insert(storage_items, cache_item)
                end
            end
            table.insert(storages, { name = storage_name, size = size, items = json_list(storage_items) })
        end
    end
    return json_list(storages)
end

local function move_stack(params)
    return m.callRemote(
        params['from']['inventoryName'],
        'pushItems',
        params['to']['inventoryName'],
        params['from']['slot'],
        params['amount'],
        params['to']['slot']
    )
end

local function get_item_detail(slot_ref)
    return m.callRemote(slot_ref['inventoryName'], 'getItemDetail', slot_ref['slot'])
end

local function get_inventory_items(storage_name)
    local storage_items = {}
    local items = m.callRemote(storage_name, 'list')
    for slot, item in pairs(items) do
        local details = {
            name = item['name'],
            nbt = item['nbt'],
            count = item['count'],
            maxCount = item['count'],
        }

        if details ~= nil then
            local cache_item = { item = details, slot = slot }
            table.insert(storage_items, cache_item)
        end
    end
    return json_list(storage_items)
end

local function get_fluid_containers(prefixes)
    local containers = {}
    peripheral.find(prefixes[1], function(name, container)
        local tanks = {}

        for slot, tank in pairs(container.tanks()) do
            table.insert(tanks, {
                slot = slot,
                fluid = tank,
            })
        end

        table.insert(containers, {
            name = name,
            tanks = json_list(tanks),
        })
        return false
    end)
    return json_list(containers)
end

local function get_tanks(name)
    local result = {}
    for i, tank in pairs(m.callRemote(name, 'tanks')) do
        table.insert(result, {
            slot = i,
            fluid = tank,
        })
    end
    return json_list(result)
end

local function move_fluid(params)
    return m.callRemote(params.from, 'pushFluid', params.to, params.amount, params.fluid)
end

local function measure_time(func)
    return function(...)
        local start_time = os.epoch 'local'
        local result = func(...)
        local end_time = os.epoch 'local'
        local elapsed_time = end_time - start_time
        print(elapsed_time)
        return result
    end
end
return function(methods, _, _)
    local cfg_load = loadfile 'config.lua'
    methods['getItems'] = measure_time(get_items)
    methods['moveStack'] = measure_time(move_stack)
    methods['getItemDetail'] = measure_time(get_item_detail)
    methods['getInventoryItems'] = measure_time(get_inventory_items)
    methods['getFluidContainers'] = get_fluid_containers
    methods['getTanks'] = get_tanks
    methods['moveFluid'] = move_fluid

    if cfg_load ~= nil then
        config = cfg_load()
    end
    return config
end
