-- Copyright 2015 The tgbot Authors. All rights reserved.
-- Use of this source code is governed by a BSD-style
-- license that can be found in the LICENSE file.

started = 0

-- We must avoid \n, \r and any other character that could break the parsing
function filter_chrs(str)
	local s = ""
	for i = 1, string.len(str) do
		if string.byte(str, i) >= 32 then
			s = s..string.sub(str, i, i)
		else
			s = s.." "
		end
	end
	return s
end

function sanitize_id(str)
	return string.gsub(filter_chrs(str), " +", "_")
end

function get_title(from, to)
	if to.type == "user" then
		return from.print_name
	elseif to.type == "chat" then
		return to.print_name
	elseif to.type == "encr_chat" then
		return from.print_name
	else
		return ""
	end
end

function on_msg_receive(msg)
	if started < 2 then -- binlog_replay_end and get_difference_end
		return
	end
	if msg.out then
		return
	end
	print("[MSG] "..
		sanitize_id(get_title(msg.from, msg.to)).." "..
		sanitize_id(msg.from.print_name).." "..
		filter_chrs(msg.text))
end

function on_binlog_replay_end()
	started = started + 1
end

function on_get_difference_end()
	started = started + 1
end

-- Fix error "*** lua: attempt to call a nil value"
function on_our_id(id) end
function on_user_update(user, what) end
function on_chat_update(chat, what) end
function on_secret_chat_update(schat, what) end
