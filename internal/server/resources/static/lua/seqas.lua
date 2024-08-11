local m = peripheral.find 'modem' --[[@as ccTweaked.peripherals.WiredModem]]

local recipes = {}
local config = {}

local function get_seq(item)
    for _, recipe in pairs(recipes) do
        if recipe['input'] == item['id'] then
            return recipe['sequence'][1]
        end
        if recipe['transitionalItem'] == item['id'] then
            local step = item['tag']['SequencedAssembly']['Step']
            local seqs = recipe['sequence']
            local idx = math.mod(step, #seqs) + 1
            return recipe['sequence'][idx]
        end
    end
    return nil
end

local function find_item(item)
    for slot, i in pairs(m.callRemote(config['buf_name'], 'list')) do
        if i['name'] == item then
            return slot
        end
    end
    return nil
end

local function sendToDeploying(slot, item)
    local reagent_slot = find_item(item)
    if reagent_slot == nil then
        print('Not found reagent: ' .. item)
        return false
    end

    local res = m.callRemote(config['buf_name'], 'pushItems', config['deployer_reagent'], reagent_slot, 1)
    if res ~= 1 then
        return true
    end

    m.callRemote(config['buf_name'], 'pushItems', config['deployer_depot'], slot, 1)
    return true
end

local function sendToPressing(slot)
    local res = m.callRemote(config['buf_name'], 'pushItems', config['press_depot'], slot, 1)
    return true
end

local function check(item)
    local seq = get_seq(item)
    if seq == nil then
        return false
    end
    local slot = item['Slot'] + 1

    print('get seq for slot: ' .. slot)

    if seq['type'] == 'deploying' then
        print('deploy to ' .. seq['input'][2])
        return sendToDeploying(slot, seq['input'][2])
    end

    if seq['type'] == 'pressing' then
        return sendToPressing(slot)
    end

    return false
end

local function step()
    local block_data = m.callRemote(config['block_reader_name'], 'getBlockData')
    local items = block_data.Items

    for _, item in pairs(items) do
        if check(item) then
            return true
        end
    end
    return false
end

local function load_config()
    local cfg_load = loadfile 'config.lua'
    if cfg_load ~= nil then
        config = cfg_load()
    end

    local f = fs.open('recipes.json', 'r')
    if f ~= nil then
        local content = f.readAll()
        if content ~= nil then
            recipes = textutils.unserializeJSON(content)
        end
        f.close()
    end
end

return function(_, handlers, _)
    load_config()

    local timerDuration = 1
    local timerId = os.startTimer(timerDuration)

    handlers['timer'] = function(eventData)
        if eventData[2] == timerId then
            m.callRemote(config['buf_name'], 'pullItems', config['deployer_reagent'], 1)
            m.callRemote(config['buf_name'], 'pullItems', config['deployer_depot'], 1)
            m.callRemote(config['buf_name'], 'pullItems', config['press_depot'], 1)

            timerDuration = 10
            if step() then
                timerDuration = 2
            end
            timerId = os.startTimer(timerDuration)
        end
    end
    return {}
end
