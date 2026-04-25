function on_startup()
    -- 起動時に CC で初期値を送る
    send_cc(0, 7, 100)   -- Ch 1 Volume = 100
    send_cc(0, 10, 64)   -- Ch 1 Pan = center
end

-- ch1 のノートを +12 に移調、ch2 はベロシティ半分
function on_note_on(ch, note, vel)
    if ch == 0 then
        send_note_on(0, note + 12, vel)
    elseif ch == 1 then
        send_note_on(1, note, math.floor(vel / 2))
    else
        send_note_on(ch, note, vel)
    end
end

function on_note_off(ch, note, vel)
    if ch == 0 then
        send_note_off(0, note + 12, vel)
    else
        send_note_off(ch, note, vel)
    end
end

-- CC1 (mod wheel) を CC11 (expression) に付け替え
function on_cc(ch, cc, val)
    if cc == 1 then
        send_cc(ch, 11, val)
    else
        send_cc(ch, cc, val)
    end
end
