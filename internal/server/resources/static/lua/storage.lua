local m = peripheral.find 'modem' --[[@as ccTweaked.peripherals.WiredModem]]


---@param inventory string
local function pull_all(inventory) 
    local items = m.callRemote(inventory, 'list')
    for slot, item in pairs(items) do
    end
end

local function import_stacks(params)

end


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

local function getItems()
    local storages = {}

    for _, storage_name in pairs(m.getNamesRemote()) do
        if is_storage(m, storage_name) then
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
            if #storage_items == 0 then
                storage_items = textutils.empty_json_array
            end
            table.insert(storages, { name = storage_name, size = size, items = storage_items })
        end
    end
    return storages
end

local function measureTime(func)
    return function()
        local start_time = os.epoch 'local'
        local result = func()
        local end_time = os.epoch 'local'
        local elapsed_time = end_time - start_time
        print(elapsed_time)
        return result
    end
end

return function(methods, _, _)
    methods['getItems'] = measureTime(getItems)
    return {}
end
