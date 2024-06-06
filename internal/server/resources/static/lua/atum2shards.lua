-- local buff = peripheral.wrap 'top' --[[@as ccTweaked.peripherals.Inventory]]
local storage = peripheral.wrap 'bottom' --[[@as ccTweaked.peripherals.Inventory]]
local furnace = peripheral.wrap 'left'--[[@as ccTweaked.peripherals.Inventory]]

local function clear_buff()
    storage.pullItems('top', 1, 64)
end

local function clean_turtle()
    for i = 1, 16 do
        if turtle.getItemCount(i) > 0 then
            turtle.select(i)
            turtle.dropUp()
            clear_buff()
        end
    end
end

local function is_need_artifact()
    local result = furnace.getItemDetail(3)
    if result ~= nil then
        furnace.pushItems('botton', 3)
    end

    local fuel = furnace.getItemDetail(2)
    if fuel == nil or fuel['name'] ~= 'minecraft:lava_bucket' then
        return false
    end

    local reagent = furnace.getItemDetail(1)
    if reagent == nil then
        return true
    end
    return false
end

local function move_item_to_buff(name)
    for slot, item in pairs(storage.list()) do
        if item['name'] == name then
            local res = storage.pushItems('top', slot, 1)
            if res == 1 then
                return true
            end
        end
    end
    return false
end

local function place_item_in_slot(slot, name)
    if not move_item_to_buff(name) then
        return false
    end
    turtle.select(slot)
    turtle.suckUp()
    return true
end

local function place_block()
    if not move_item_to_buff 'atum:godforged_block' then
        return false
    end
    clean_turtle()
    turtle.suckUp()
    turtle.place()
    clean_turtle()
    return true
end

local function check_ingredients(ings)
    for _, item in pairs(storage.list()) do
        if ings[item['name']] ~= nil then
            ings[item['name']] = ings[item['name']] - item['count']
        end
    end
    for _, count in pairs(ings) do
        if count > 0 then
            return false
        end
    end
    return true
end

local function craft_block(shard)
    local comb = 'resourcefulbees:dusty_mummbee_honeycomb'
    local nebu = 'atum:nebu_ingot'

    clean_turtle()
    if not check_ingredients {
        [comb] = 3,
        [nebu] = 4,
        [shard] = 2,
    } then
        return false
    end

    local reagent = {
        [1] = nebu,
        [2] = comb,
        [3] = nebu,
        [5] = shard,
        [6] = comb,
        [7] = shard,
        [9] = nebu,
        [10] = comb,
        [11] = nebu,
    }

    for slot, name in pairs(reagent) do
        if not place_item_in_slot(slot, name) then
            print('cant place ' .. name .. ' at ' .. slot)
            return false
        end
    end

    if not turtle.craft(1) then
        clean_turtle()
        return false
    end
    return true
end

local function endswith(str, suffix)
    if string.len(str) < string.len(suffix) then
        return false
    end

    return string.sub(str, string.len(str) - string.len(suffix) + 1, string.len(str)) == suffix
end

local function select_shard()
    local counts = {}
    for _, item in pairs(storage.list()) do
        local name = item['name']
        if endswith(name, '_godshard') then
            if counts[name] == nil then
                counts[name] = 0
            end
            counts[name] = counts[name] + item['count']
        end
    end
    local max_name = nil
    local max_count = 0
    for name, count in pairs(counts) do
        if count > max_count then
            max_count = count
            max_name = name
        end
    end
    return max_name
end

local function check_state()
    if not is_need_artifact() then
        print 'no need artifact'
        return false
    end

    if turtle.inspect() then
        print 'has block'
        return false
    end

    if place_block() then
        print 'block placed'
        return false
    end

    local shard = select_shard()
    if shard == nil then
        print "can't find shard"
        return false
    end

    if not craft_block(shard) then
        print "can't creaete block"
        return false
    end
    clean_turtle()
    if not place_block() then
        return false
    end

    print 'OK'
    return true
end

local timer_id = os.startTimer(60)
while true do
    local _, id = os.pullEvent 'timer'
    if timer_id == id then
        check_state()
        timer_id = os.startTimer(60)
    end
end
