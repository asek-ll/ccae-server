local glass = {
    name = 'minecraft:glass',
    count = 1,
    maxCount = 64,
}
testnet.registerRecipe {
    result = {
        name = 'minecraft:glass_pane',
        count = 16,
        maxCount = 64,
    },
    ingredients = {
        [1] = glass,
        [2] = glass,
        [3] = glass,
        [4] = glass,
        [5] = glass,
        [6] = glass,
    },
}

local network = 'net1'

local m1 = testnet.createModem(network)

local i1 = testnet.createInventory(network, 'inventory_cold_1')
local i2 = testnet.createInventory(network, 'inventory_warm_1')
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

ccemux.openEmu(1)
testnet.attachPeripheral('top', m1, 1)

local crafter = 2
testnet.enableTurtleApi(crafter)
local crafter_modem = testnet.createModem(network)
local crafter_input = testnet.createInventory(network, 'crafter_input')
local crafter_output = testnet.createInventory(network, 'crafter_output')
local crafter_buffer = testnet.createInventory(network, 'crafter_buffer')
-- testnet.setItem(crafter_buffer, 1, {
--     name = 'resourcefulbees:bee_jar',
--     count = 1,
--     maxCount = 64,
-- })
ccemux.openEmu(crafter)
testnet.attachPeripheral('top', crafter_input, crafter)
testnet.attachPeripheral('bottom', crafter_buffer, crafter)
testnet.attachPeripheral('front', crafter_output, crafter)
testnet.attachPeripheral('left', crafter_modem, crafter)
testnet.attachWorkbench('right', crafter)

-- print(crafter_buffer)
-- print(i2)
-- print(storage_input)

ccemux.closeEmu()
