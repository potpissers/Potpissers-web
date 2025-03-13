package main

const ReturnServerTips = `SELECT tip_message
FROM server_tips
WHERE game_mode_name = $1`
const Return12Deaths = `SELECT name,
       victim_user_fight_id,
       timestamp,
       victim_uuid,
       bukkit_victim_inventory,
       death_world,
       death_x,
       death_y,
       death_z,
       death_message,
       killer_uuid,
       bukkit_kill_weapon,
       bukkit_killer_inventory
FROM user_deaths
         JOIN servers ON user_deaths.server_id = servers.id
ORDER BY timestamp DESC
LIMIT 12`
const Return12ServerDeaths = `SELECT name,
       victim_user_fight_id,
       timestamp,
       victim_uuid,
       bukkit_victim_inventory,
       death_world,
       death_x,
       death_y,
       death_z,
       death_message,
       killer_uuid,
       bukkit_kill_weapon,
       bukkit_killer_inventory
FROM user_deaths
         JOIN servers ON user_deaths.server_id = servers.id
WHERE server_id = (SELECT id FROM servers WHERE name = $1)
ORDER BY timestamp DESC
LIMIT 12`
const Return12NewPlayers = `SELECT user_uuid, referrer, timestamp, ROW_NUMBER() OVER (ORDER BY timestamp) AS row_number
FROM user_referrals
ORDER BY timestamp
LIMIT 12`
const Return14Events = `SELECT start_timestamp,
       loot_factor,
       max_timer,
       is_movement_restricted,
       CASE WHEN end_timestamp IS NOT NULL THEN capping_user_uuid END AS capping_user_uuid,
       end_timestamp,
       capping_party_uuid,
       world,
       x,
       y,
       z,
       servers.name                                                   AS server_name,
       arena_data.name                                                AS arena_name,
       creator
FROM koths
         JOIN server_koths ON server_koths_id = server_koths.id
         JOIN servers ON servers.id = server_koths.server_id
         JOIN arena_data ON arena_data.id = server_koths.arena_id
ORDER BY end_timestamp IS NULL, end_timestamp
LIMIT 14`
const Return14ServerEvents = `SELECT start_timestamp,
       loot_factor,
       max_timer,
       is_movement_restricted,
       CASE WHEN end_timestamp IS NOT NULL THEN capping_user_uuid END AS capping_user_uuid,
       end_timestamp,
       capping_party_uuid,
       world,
       x,
       y,
       z,
       servers.name                                                   AS server_name,
       arena_data.name                                                AS arena_name,
       creator
FROM koths
         JOIN server_koths ON server_koths_id = server_koths.id
         JOIN servers ON servers.id = server_koths.server_id
         JOIN arena_data ON arena_data.id = server_koths.arena_id
WHERE server_id = (SELECT id FROM servers WHERE name = $1)
ORDER BY end_timestamp IS NULL, end_timestamp
LIMIT 14`
const ReturnAllServerData = `SELECT death_ban_minutes,
       world_border_radius,
       sharpness_limit,
       power_limit,
       protection_limit,
       bard_regen_level,
       bard_strength_level,
       is_weakness_enabled,
       is_bard_passive_debuffing_enabled,
       dtr_freeze_timer,
       dtr_max,
       dtr_max_time,
       dtr_off_peak_freeze_time,
       off_peak_lives_needed_as_cents,
       bard_radius,
       rogue_radius,
       servers.name,
       attack_speeds.name
FROM server_data
         JOIN servers ON id = server_id
         JOIN attack_speeds ON attack_speed_id = attack_speeds.id`
const ReturnAllOnlinePlayers = `SELECT user_name, name
FROM online_players
         JOIN servers ON server_id = servers.id`
const Return7ServerFactions = `SELECT name, party_uuid
FROM factions
WHERE server_id = (SELECT id FROM servers WHERE name = $1)
LIMIT 7`
const Return7ServerBandits = `SELECT user_uuid,
       death_id,
       timestamp,
       expiration_timestamp,
       death_message,
       death_world,
       death_x,
       death_y,
       death_z
FROM bandits
         JOIN user_deaths on bandits.death_id = user_deaths.id
WHERE bandits.server_id = (SELECT id FROM servers WHERE name = $1)
  AND expiration_timestamp > NOW()
ORDER BY timestamp DESC
LIMIT 7`
const ReturnUnsuccessfulTransactions = `SELECT order_id
FROM unnest($1::TEXT[]) AS order_id
WHERE order_id NOT IN (SELECT square_order_id
                       FROM successful_transactions)`
const InsertSuccessfulTransaction = `INSERT INTO successful_transactions (square_order_id, user_uuid, line_item_id, line_item_player_name,
                                     line_item_quantity,
                                     amount_as_cents, referrer)
VALUES ($1, $2, (SELECT id FROM line_items WHERE line_item_name = $3), $4, $5, $6,
        (SELECT referrer FROM user_referrals WHERE user_uuid = $2))`

const ReturnAllLineItems = `SELECT game_mode_name, line_item_name, value_in_cents, description, is_plural
FROM line_items
ORDER BY id`
