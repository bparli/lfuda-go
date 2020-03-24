package lfuda

import (
	"math/rand"
	"testing"
)

func BenchmarkLFUDA(b *testing.B) {
	l := New(8192)

	trace := make([]int64, b.N*2)
	for i := 0; i < b.N*2; i++ {
		if i%2 == 0 {
			trace[i] = rand.Int63() % 16384
		} else {
			trace[i] = rand.Int63() % 32768
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.Set(trace[i], trace[i])
	}
	var hit, miss int
	for i := 0; i < b.N; i++ {
		_, ok := l.Get(trace[i])
		if ok {
			hit++
		} else {
			miss++
		}
	}
	b.Logf("hit: %d miss: %d ratio: %f", hit, miss, float64(hit)/float64(miss))
}

func BenchmarkLFUDA_Rand(b *testing.B) {
	l := New(8192)

	trace := make([]int64, b.N*2)
	for i := 0; i < b.N*2; i++ {
		trace[i] = rand.Int63() % 32768
	}

	b.ResetTimer()

	var hit, miss int
	for i := 0; i < 2*b.N; i++ {
		if i%2 == 0 {
			l.Set(trace[i], trace[i])
		}
		if i%7 == 0 {
			for j := 0; j < 20; j++ {
				_, ok := l.Get(trace[i])
				if ok {
					hit++
				} else {
					miss++
				}
			}
		} else {
			_, ok := l.Get(trace[i])
			if ok {
				hit++
			} else {
				miss++
			}
		}
	}
	b.Logf("hit: %d miss: %d ratio: %f", hit, miss, float64(hit)/float64(miss))
}

func TestLFUDA(t *testing.T) {
	evictCounter := 0
	onEvicted := func(k interface{}, v interface{}) {
		if k != v {
			t.Errorf("Evict values not equal (%v!=%v)", k, v)
		}
		evictCounter++
	}
	l := NewWithEvict(666, onEvicted)

	numSet := 0
	for i := 100; i < 1000; i++ {
		if l.Set(i, i) {
			numSet++
		}
	}
	if l.Len() != 222 || l.Len() != len(l.Keys()) {
		t.Errorf("bad len: %v", l.Len())
	}

	if evictCounter != 900-222 {
		t.Errorf("bad evict count: %v", evictCounter)
	}

	// TEST [988 989 990 991 992 993 994 995 996 997 998 999 766 767 768 769 770 771 772 773 774 775 776 777 778 779 780 781 782 783 784 785 786 787 788 789 790 791 792 793 794 795 796 797 798 799 800 801 802 803 804 805 806 807 808 809 810 811 812 813 814 815 816 817 818 819 820 821 822 823 824 825 826 827 828 829 830 831 832 833 834 835 836 837 838 839 840 841 842 843 844 845 846 847 848 849 850 851 852 853 854 855 856 857 858 859 860 861 862 863 864 865 866 867 868 869 870 871 872 873 874 875 876 877 878 879 880 881 882 883 884 885 886 887 888 889 890 891 892 893 894 895 896 897 898 899 900 901 902 903 904 905 906 907 908 909 910 911 912 913 914 915 916 917 918 919 920 921 922 923 924 925 926 927 928 929 930 931 932 933 934 935 936 937 938 939 940 941 942 943 944 945 946 947 948 949 950 951 952 953 954 955 956 957 958 959 960 961 962 963 964 965 966 967 968 969 970 971 972 973 974 975]

	// fmt.Println("TEST", l.Keys(), l.Age())
	// fmt.Println("TEST", l.Age())
	for _, k := range l.Keys() {
		if v, ok := l.Get(k); !ok || v != k {
			t.Fatalf("bad key: %v, %v, %t", k, v, ok)
		}

	}

	// these should all be misses since their hits will be too low
	// relative to newer keys set when the cache is more aged
	for i := 100; i < 765; i++ {
		_, ok := l.Get(i)
		if ok {
			t.Fatalf("should not be in cache")
		}
	}

	if ok := l.Set(256, 256); !ok {
		t.Errorf("Wasn't able to set key/value in cache (but should have been)")
	}

	if val, _ := l.Get(256); val != 256 {
		t.Errorf("Wrong value returned for key")
	}

	// expect 256 to be the top hit in l.Keys()
	if l.Keys()[0] != 256 {
		t.Errorf("out of order key (last set of keys should have hits=5 and should be first): %d", l.Keys()[0])
	}

	l.Purge()
	if l.Len() != 0 {
		t.Errorf("bad len: %v", l.Len())
	}
	if _, ok := l.Get(200); ok {
		t.Errorf("should contain nothing")
	}
}

// test that Set returns true/false
func TestLFUDASet(t *testing.T) {
	evictCounter := 0
	onEvicted := func(k interface{}, v interface{}) {
		evictCounter++
	}

	l := NewWithEvict(1, onEvicted)

	if l.Set(1, 1) == true || evictCounter != 0 {
		t.Errorf("should not have evicted")
	}
	if l.Set(2, 2) == false || evictCounter != 1 {
		t.Errorf("should have evicted")
	}
}

// test that Contains doesn't update recent-ness
func TestLFUDAContains(t *testing.T) {
	l := NewWithEvict(2, nil)
	l.Set(1, 1)
	l.Set(2, 2)

	// bump 1 hits
	for i := 0; i < 10; i++ {
		l.Get(1)
	}

	if l.Keys()[0] != 1 {
		t.Errorf("key 1 should be the most frequently used")
	}

	// should not bump 2 hits
	for i := 0; i < 20; i++ {
		l.Contains(2)
	}

	if l.Keys()[0] != 1 {
		t.Errorf("key 1 should be the most frequently used")
	}
}

// test that ContainsOrSet doesn't update recent-ness
func TestLFUDAContainsOrSet(t *testing.T) {
	l := New(2)

	l.Set(1, 1)
	l.Set(2, 2)
	contains, eviction := l.ContainsOrSet(1, 1)
	if !contains {
		t.Errorf("1 should be contained")
	}
	if eviction {
		t.Errorf("nothing should have been set")
	}

	contains, eviction = l.ContainsOrSet(3, 3)
	if contains {
		t.Errorf("3 should not have been contained")
	}
	if !eviction {
		t.Errorf("3 should have been set and an eviction should have occurred")
	}
}

// test that PeekOrSet doesn't update recent-ness
func TestLFUDAPeekOrSet(t *testing.T) {
	l := New(2)

	l.Set(1, 1)
	l.Set(2, 2)
	previous, contains, set := l.PeekOrSet(1, 1)
	if !contains {
		t.Errorf("1 should be contained")
	}
	if set {
		t.Errorf("nothing should have been set here")
	}
	if previous != 1 {
		t.Errorf("previous is not equal to 1")
	}

	_, contains, set = l.PeekOrSet(3, 3)
	if contains {
		t.Errorf("3 should not have been contained")
	}
	if !set {
		t.Errorf("3 should have been set here")
	}

	l.Get(3)
	_, contains, set = l.PeekOrSet(3, 3)
	if !contains {
		t.Errorf("3 should have been contained")
	}
	if set {
		t.Errorf("3 should not have been set")
	}

	previous, contains, set = l.PeekOrSet(3, 3)
	if previous != 3 {
		t.Errorf("3 should be returned")
	}
	if !contains {
		t.Errorf("3 should have been contained")
	}
	if set {
		t.Errorf("nothing should be set here")
	}
}

// test that Peek doesn't update recent-ness
func TestLFUDAPeek(t *testing.T) {
	l := New(2)

	l.Set(1, 1)
	l.Set(2, 2)
	if v, ok := l.Peek(1); !ok || v != 1 {
		t.Errorf("1 should be set to 1: %v, %v", v, ok)
	}

	l.Get(2)
	l.Set(3, 3)
	if l.Contains(1) {
		t.Errorf("should not have updated hits for 1")
	}
}

func TestLFUDARemove(t *testing.T) {
	l := New(2)

	l.Set(1, 1)
	l.Set(2, 2)
	if v, ok := l.Get(1); !ok || v != 1 {
		t.Errorf("1 should be set to 1: %v, %v", v, ok)
	}

	l.Remove(1)
	if _, ok := l.Get(1); ok {
		t.Errorf("1 should not be in the cache")
	}
	if l.Len() != 1 {
		t.Errorf("Cache should be length 1 (but it wasn't)")
	}
}

func TestLFUDAAge(t *testing.T) {
	l := New(1)

	l.Set(1, 1)

	// bump hits on key 1 to 2
	l.Get(1)
	if evicted := l.Set(2, 2); !evicted {
		t.Errorf("Set op should have evicted (but it didn't)")
	}

	// key 1 was evicted so cache age should be its hits value (2)
	if l.Age() != 2 {
		t.Errorf("cache age should have been set to 1 (but it was't)")
	}
}

func TestLFUDASize(t *testing.T) {
	l := New(11)

	for i := 10; i < 30; i++ {
		l.Set(i, i)
	}

	if l.Size() != 10 {
		t.Errorf("Cache can only store up to 11 bytes so Size should be divisible by 2")
	}

	l.Purge()
	if l.Size() != 0 {
		t.Errorf("Cache size should be reset to 0 (but it wasn't)")
	}
}
