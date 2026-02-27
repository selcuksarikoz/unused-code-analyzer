# feature.rb
require_relative 'utils'

module Feature
  def self.run
    Utils.used_function
  end
  
  def self.unused_method
    "unused"
  end
end

class FeatureClass
  include Utils
  
  def initialize
    @value = Utils.used_function
  end
  
  def used_method
    @value
  end
  
  def unused_method
    "unused"
  end
end

def used_ruby_function(used_param)
  puts used_param
end

def unused_ruby_function(unused1, unused2)
  "unused"
end

unused_local = "this is unused"

def local_method
end
