def pluck_one(l):
    if (type(l) == list or type(l) == tuple) and len(l) == 1:
        return l[0]
    else:
        return l
