local m = peripheral.find 'modem' --[[@as ccTweaked.peripherals.WiredModem]]

local config = {}

local function check()
    local list = m.callRemote(config['energizer'], 'list')
    local result = list[1]
    if result ~= nil then
        if m.callRemote(config['energizer'], 'pushItems', config['output'], 1) ~= result['count'] then
            print 'cant move result out'
            return false
        end
    end

    local input_list = m.callRemote(config['input'], 'list')

    for slot in pairs(input_list) do
        local eslot = slot + 1
        if list[eslot] == nil then
            if m.callRemote(config['input'], 'pushItems', config['energizer'], slot, 1, eslot) == 0 then
                print "can't move ing in"
                return false
            end
        end
    end

    print 'ok'
    return true
end

local function load_config()
    local cfg_load = loadfile 'config.lua'
    if cfg_load ~= nil then
        config = cfg_load()
    end
end

return function(_, handlers, _)
    load_config()

    local timerDuration = 1
    local timerId = os.startTimer(timerDuration)

    handlers['timer'] = function(eventData)
        if eventData[2] == timerId then
            timerDuration = 10
            check()
            timerId = os.startTimer(timerDuration)
        end
    end
    return {}
end
