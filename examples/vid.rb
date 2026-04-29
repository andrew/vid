#!/usr/bin/env ruby
require 'digest'

B32 = [*'a'..'z', *'2'..'7'].freeze

def gitoid(path)
  bytes = File.binread(path)
  Digest::SHA256.hexdigest("blob #{bytes.bytesize}\0" + bytes)
end

def base32_120(bytes)
  bytes.unpack1('B*').chars.each_slice(5).map { |g| B32[g.join.to_i(2)] }.join
end

def vid(purl, path)
  digest = Digest::SHA256.digest("#{purl}\n#{gitoid(path)}")
  'VID-1-' + base32_120(digest[0, 15]).scan(/.{4}/).join('-')
end

if ARGV.length != 2
  warn "usage: #{$PROGRAM_NAME} <purl> <file>"
  exit 1
end

puts vid(ARGV[0], ARGV[1])
