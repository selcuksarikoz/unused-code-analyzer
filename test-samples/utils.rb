# utils.rb
module Utils
  def self.used_function
    "hello"
  end
  
  USED_CONST = "world"
  
  class UsedClass
    def initialize(name)
      @name = name
    end
  end
  
  def self.unused_function
    "unused"
  end
  
  UNUSED_CONST = 123
  
  class UnusedClass
  end
end
