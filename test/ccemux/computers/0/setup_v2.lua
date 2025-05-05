return function(network, computer_id)
    local m1 = testnet.createModem(network)


    ccemux.openEmu(computer_id)
    testnet.attachPeripheral('top', m1, computer_id)
end
