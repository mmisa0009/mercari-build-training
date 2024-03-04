
class exercise_wordpattern:
    def wordPattern(self, pattern: str, s: str) -> bool:
        words = s.split()
        word_dict = {}
        word_invented_dict = {}

        if len (words) != len(pattern):
            return False
        for i, p in enumerate(pattern):
            if p not in word_dict and words[i] not in word_invented_dict:
                word_dict[p] = words[i]
                word_invented_dict[words[i]] = p
            elif p in word_dict and words[i] in word_invented_dict:
                if word_dict[p] != words[i]:
                    return False
            else:
                return False
        
        return True
    

#test 
if __name__ == "__main__":
    solution = exercise_wordpattern()
    pattern = "abba"
    s = "cat dog dog cat"
    print(solution.wordPattern(pattern, s))


class exercise_array:
    def find_numbers(self, nums):
        result = []
        for num in nums:
            index = abs(num) -1
            if nums[index]>0:
                nums[index] = -nums[index]
            
        for i, num in enumerate(nums):
            if num >0:
                result.append (i+1)
        return result
    
class ListNode:
    def __init__ (self, val=0, next=None):
        self.val = val
        self.next = next

    def getIntersectionNode(headA, headB):
        def getLength(head):
            length = 0
            while head:
                length += 1
                head = head.next
            return length
        
        lenA = getLength(headA)
        lenB = getLength(headB)

        diff = abs(lenA-lenB)

        if lenA > lenB:
            for _ in range(diff):
                headA = headA.next
            else:
                for _ in range(diff):
                    headB = headB.next

        while headA and headB:
            if headA == headB:
                return headA
            headA = headA.next
            headB = headB.next

        return None
    
#test
if __name__ == "__main__":
    intersectNode = ListNode(3, ListNode(4, ListNode(5) ))
    headA = ListNode(1, ListNode(2, intersectNode))
    headB = ListNode(7, ListNode(8, intersectNode))

    intersection = ListNode.getIntersectionNode (headA, headB)
    if intersection:
        print("Intersection node value:", intersection.val)
    else:
        print("No intersection found")

    headC = ListNode(6)
    headD = ListNode(9)
    noIntersection = ListNode.getIntersectionNode(headC, headD)
    if noIntersection:
        print("Intersection node value:", noIntersection.val)
    else:
        print("No intersection found")