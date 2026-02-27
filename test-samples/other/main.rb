# main.rb
require_relative 'utils'
require 'nonexistent_gem'

puts Utils.used_function
puts Utils::USED_CONST

obj = Utils::UsedClass.new("test")
puts obj

include Utils
puts used_function

extend Utils

class MyClass
  include Utils
end

unused_local = "this is unused"

def local_method
end

class LocalClass
end
