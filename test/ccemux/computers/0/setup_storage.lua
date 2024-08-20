return function(network, computer_id)
    local m1 = testnet.createModem(network)

    local i1 = testnet.createInventory(network, 'inventory_cold_1')
    local i2 = testnet.createInventory(network, 'inventory_warm_1')
    local f1 = testnet.createTank(network, {
        type = 'storage_tank',
    })
    testnet.setFluid(f1, {
        name = 'minecraft:water',
        amount = 500,
    })
    local f2 = testnet.createTank(network, {
        type = 'storage_tank',
    })

    local storage_input = testnet.createInventory(network, 'storage_input')

    testnet.setItem(i1, 1, {
        name = 'minecraft:apple',
        count = 2,
        maxCount = 64,
    })

    testnet.setItem(i2, 2, {
        name = 'minecraft:ender_pearl',
        count = 3,
        maxCount = 16,
    })

    testnet.setItem(i2, 4, {
        name = 'pneumaticcraft:advanced_pcb',
        count = 30,
        maxCount = 64,
    })

    testnet.setItem(i2, 5, {
        name = 'pneumaticcraft:printed_circuit_board',
        count = 60,
        maxCount = 64,
    })

    testnet.setItem(i2, 6, {
        name = 'minecraft:glass',
        count = 64,
        maxCount = 64,
    })

    testnet.setItem(i2, 7, {
        name = 'emendatusenigmatica:bronze_ingot',
        count = 31,
        maxCount = 64,
    })

    testnet.setItem(i2, 8, {
        name = 'emendatusenigmatica:copper_ingot',
        count = 20,
        maxCount = 64,
    })

    testnet.setItem(i2, 6, {
        name = 'minecraft:glass',
        count = 58,
        maxCount = 64,
    })

    testnet.setItem(i2, 7, {
        name = 'minecraft:glass',
        count = 58,
        maxCount = 64,
    })

    testnet.setItem(i2, 8, {
        name = 'minecraft:glass',
        count = 58,
        maxCount = 64,
    })

    local trans_storage = testnet.createInventory(network, 'transaction_storage')
    local trans_tank = testnet.createTank(network, { name = 'transaction_tank' })

    ccemux.openEmu(computer_id)
    testnet.attachPeripheral('top', m1, computer_id)
end
