function on_startup()
end

function on_note_on(ch, note, vel)
    send_note_on(ch, note, vel)
end

function on_note_off(ch, note, vel)
    send_note_off(ch, note, vel)
end

function on_cc(ch, cc, val)
    send_cc(ch, cc, val)
end
