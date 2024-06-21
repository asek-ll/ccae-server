
return function(methods, handlers, wsclient)
    config = setup()
    methods['dumpOut'] = dump_out
    methods['craft'] = craft
    methods['restore'] = restore
    methods['processResults'] = processResults
    return config
end
