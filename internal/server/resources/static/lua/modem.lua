local m = peripheral.find 'modem' --[[@as ccTweaked.peripherals.WiredModem]]

return function(methods, _, _)
    for n, f in pairs(m) do
        methods[n] = f
    end

    return {}
end
